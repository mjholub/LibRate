<script lang="ts">
	import axios from 'axios';
	import { onDestroy } from 'svelte';
	import { browser } from '$app/environment';
	import { authStore } from '../../stores/members/auth.ts';
	import PasswordInput from './PasswordInput.svelte';
	import { PasswordMeter } from 'password-meter';

	const tooltipMessage = 'Not recommended on shared computers';
	let isRegistration = false;
	let email_or_username = '';
	if (browser) {
		email_or_username = localStorage.getItem('email_or_username') || '';
	}
	let email = '';
	let nickname = '';

	let isEmailAvailable = true;
	let isNickAvailable = true;

	let password = '';
	let rememberMe = false;
	let showPassword = false;
	let passwordConfirm = '';
	let passwordStrength = '' as string; // it is based on the message from the backend, not the entropy score
	let errorMessage = '';
	let strength: number;
	let email_input: HTMLInputElement;
	let nickname_input: HTMLInputElement;

	const toggleObfuscation = () => {
		showPassword = !showPassword;
	};

	// helper function to check password strength
	let timeoutId: number | undefined;

	const prepareCSRF = async () => {
		const csrfToken = document.cookie
			.split('; ')
			.find((row) => row.startsWith('csrf_'))
			?.split('=')[1];
		const headers = {
			'Content-Type': 'application/json',
			'X-CSRF-Token': csrfToken
		};
		return headers;
	};

	const checkEmailExistApi = async (email: string) => {
		let available = false;
		const headers = await prepareCSRF();
		const requestPayload = {
			email
		};
		const res = await axios.post('/api/members/check', requestPayload, { headers });
		res.data.message === 'available' ? (available = true) : (available = false);
		return available;
	};

	const checkNicknameExistApi = async (nickname: string) => {
		let available = false;
		const headers = await prepareCSRF();
		const requestPayload = {
			memberName: nickname
		};
		const res = await axios.post('/api/members/check', requestPayload, { headers });
		res.data.message === 'available' ? (available = true) : (available = false);
		return available;
	};

	const checkEmailExists = async (email: string, debounceTime: number) => {
		if (timeoutId) {
			window.clearTimeout(timeoutId);
		}
		// wait until there is a value in the input field
		if (email.length < 1) return;

		timeoutId = window.setTimeout(async () => {
			try {
				isEmailAvailable = await checkEmailExistApi(email);
			} catch (error) {
				isEmailAvailable = false;
			}
		}, debounceTime);
	};

	const checkNicknameExists = async (nickname: string, debounceTime: number) => {
		if (timeoutId) {
			window.clearTimeout(timeoutId);
		}

		// wait until there is a value in the input field
		if (nickname.length < 1) return;

		timeoutId = window.setTimeout(async () => {
			try {
				isNickAvailable = await checkNicknameExistApi(nickname);
			} catch (error) {
				isNickAvailable = false;
			}
		}, debounceTime);
	};

	const checkEntropy = async (password: string) => {
		// if just logging in, don't check the entropy
		if (!isRegistration) return;

		if (timeoutId) {
			window.clearTimeout(timeoutId);
		}

		timeoutId = window.setTimeout(async () => {
			try {
				strength = new PasswordMeter().getResult(password).score;
				passwordStrength = strength > 135 ? 'Password is strong enough' : `${strength / 2.9} bits`;
			} catch (error) {
				process.env.NODE_ENV === 'development'
					? console.error(error)
					: console.error('Error checking password entropy');
			}
		}, 300);
	};
	$: {
		if (isRegistration && password) {
			checkEntropy(password);
		}
	}

	const setRememberMe = (event: Event) => {
		if (browser) {
			rememberMe = (event.target as HTMLInputElement).checked;
		}
	};

	// helper function to trigger moving either email or nickname to a dedicated field
	const startRegistration = () => {
		isRegistration = true;
		if (browser) {
			email_or_username.includes('@')
				? ((email = email_or_username), localStorage.setItem('email_or_username', ''))
				: ((nickname = email_or_username), localStorage.setItem('email_or_username', ''));
		}
	};

	const register = async (event: Event) => {
		return new Promise(async (resolve, reject) => {
			let csrfToken: string | undefined;
			if (browser) {
				csrfToken = document.cookie
					.split('; ')
					.find((row) => row.startsWith('csrf_'))
					?.split('=')[1];
			}
			event.preventDefault();
			const headers = {
				'Content-Type': 'multipart/form-data',
				'Referrer-Policy': 'no-referrer-when-downgrade',
				'X-CSRF-Token': csrfToken
			};

			localStorage.removeItem('email_or_username');

			// check if passwords match when the registration flow has been triggered
			isRegistration && password !== passwordConfirm
				? ((errorMessage = 'Passwords do not match'), false)
				: passwordStrength !== 'Password is strong enough'
				? ((errorMessage = 'Password is not strong enough'), false)
				: true;

			let requestPayload = {
				membername: nickname,
				email,
				password,
				passwordConfirm,
				roles: ['regular']
			};

			const response = await axios.post('/api/authenticate/register', requestPayload, { headers });

			const { data } = response;

			if (browser) {
				if (response.data.message.includes('already taken')) {
					errorMessage = response.data.message;
					reject(errorMessage);
				}

				if (response.status == 200) {
					authStore.set({
						isAuthenticated: true
					});
					localStorage.removeItem('email_or_username');
					window.location.reload();
					console.info('Registration successful');
					resolve(data.message);
				} else {
					errorMessage = data.message;
					reject(data.message);
					console.error(data.message);
				}
			}
		});
	};
	//
	// END OF REGISTRATION
	// START OF LOGIN
	//
	const login = async (event: Event) => {
		let csrfToken: string | undefined;
		if (browser) {
			csrfToken = document.cookie
				.split('; ')
				.find((row) => row.startsWith('csrf_'))
				?.split('=')[1];
		}
		event.preventDefault();
		const headers = {
			'Content-Type': 'multipart/form-data',
			'Referrer-Policy': 'no-referrer-when-downgrade',
			'X-CSRF-Token': csrfToken
		};

		localStorage.removeItem('member');

		const nickName = email_or_username.includes('@') ? '' : email_or_username;
		const emailValue = email_or_username.includes('@') ? email_or_username : '';

		const response = await axios.postForm(
			'/api/authenticate/login',
			{
				membername: nickName,
				email: emailValue,
				remember_me: rememberMe,
				password
			},
			{
				headers: headers
			}
		);

		if (browser) {
			response.status == 200
				? (authStore.set({
						isAuthenticated: true
				  }),
				  console.debug('authStore updated to ', authStore),
				  localStorage.removeItem('email_or_username'),
				  localStorage.setItem('jwtToken', response.data.token),
				  window.location.reload(),
				  console.info('Login successful'))
				: (errorMessage = response.data.message);
			console.error(response.data.message);
		}
	};

	onDestroy(() => {
		if (browser) {
			localStorage.removeItem('email_or_username');
		}
	});
</script>

<!-- Form submission handler -->
<form on:submit|preventDefault={isRegistration ? register : login}>
	<div class="input">
		{#if isRegistration}
			<label for="email">Email:</label>
			<input
				bind:this={email_input}
				bind:value={email}
				on:blur={() => checkEmailExists(email, 1000)}
				type="email"
				class="input"
				id="email_input"
				required
				aria-label="Email"
			/>
			{#if email.length > 0}
				{#if isEmailAvailable && email_input.validity.valid}
					<p>Email is available</p>
				{:else if !email.match(/(\w+)@(\w+)\.(\w+)/)}
					<span class="spinner" />
				{:else}
					<p class="error-message">
						Email is not available. Try <a href="https://librate.fediverse.observer/"
							>another instance</a
						>
						or <a href="/form/account/recover">recover your password</a>
					</p>
				{/if}
			{/if}
			<label for="nickname">Nickname:</label>
			<input
				type="text"
				bind:this={nickname_input}
				bind:value={nickname}
				on:blur={() => checkNicknameExists(nickname, 1100)}
				required
				class="input"
				id="nickname_input"
			/>
			{#if nickname.length > 0}
				{#if isNickAvailable}
					<p class="info-message">Nickname available</p>
				{:else}
					<p class="error-message">
						Nickname not available. Try
						<a href="https://librate.fediverse.observer/">another instance</a>
						or <a href="/form/account/recover">recover your account</a>
					</p>
				{/if}
			{/if}
			<label for="password">Password:</label>
			<PasswordInput
				bind:value={password}
				id="password"
				onInput={async () => {
					checkEntropy(password);
				}}
				{showPassword}
				{toggleObfuscation}
			/>
		{:else}
			<label for="email_or_username">Email or Username:</label>
			<input
				type="text"
				id="email_or_username"
				bind:value={email_or_username}
				required
				class="input"
			/>

			<label for="password">Password:</label>
			<PasswordInput
				bind:value={password}
				id="password"
				onInput={async () => void 0}
				{showPassword}
				{toggleObfuscation}
			/>
			<label for="rememberMe"
				>Remember me<span class="tooltip" aria-label={tooltipMessage}> *</span></label
			>
			<input
				type="checkbox"
				id="rememberMe"
				name="rememberMe"
				value="rememberMe"
				on:change={setRememberMe}
			/>

			<!-- FIXME: this is not getting updated properly -->
			<!--
			{#if  isAvailable}
				<p>Nickname is available</p>
			{:else}
				<p>Nickname is not available</p>
			{/if}
      -->
		{/if}

		{#if isRegistration}
			<label for="password">Confirm Password:</label>
			<PasswordInput
				id="password"
				bind:value={passwordConfirm}
				onInput={() => Promise.resolve(void 0)}
				{showPassword}
				{toggleObfuscation}
			/>
			<!-- Password strength indicator -->
			{#if passwordStrength !== 'Password is strong enough'}
				<p>
					Password strength: {passwordStrength} of (<a
						href="https://www.omnicalculator.com/other/password-entropy">entropy</a
					>), required: 50
				</p>
			{:else}
				<p>Password strength: {passwordStrength}</p>
			{/if}
			{#if errorMessage}
				<p><span class="error-icon" /><span class="error-message">{errorMessage}</span></p>
			{/if}
		{/if}

		{#if errorMessage}
			<p><span class="error-icon" /><span class="error-message">{errorMessage}</span></p>
		{/if}
	</div>
	<!-- End of input container -->
	<div class="button-container">
		{#if !isRegistration}
			<button type="submit" on:click={login}>Sign In</button>
			<button type="button" on:click={startRegistration}>Sign Up</button>
		{:else}
			<button type="button" on:click={() => (isRegistration = false)}>Sign In</button>
			<button type="submit">Sign Up</button>
		{/if}
	</div>
</form>

<style>
	:root {
		--error-color: red;
		--error-background: #ffe6e6;
	}

	.input {
		font-family: inherit;
		font-size: inherit;
		display: inline-table;
		padding: 0.2em 0.4em;
		margin: 0.1em 0 0.1em 0;
		box-sizing: border-box;
		border: 1px solid #ccc;
		border-radius: 4px;
		width: calc(98% - 0.2em);
		height: 2rem;
		left: 0.2em;
	}

	.error-message {
		color: var(--error-color);
		background-color: var(--error-background);
		padding: 0.1rem;
		margin: 0.2em 0;
		width: inherit;
		border-radius: 4px;
		border: 1px solid var(--error-color);
		font-weight: bold;
		font-size: 0.8em;
	}

	p.info-message {
		font-size: 0.75em;
		margin: 0.2em 0;
		word-break: break-word;
		word-wrap: break-word;
	}

	.error-icon::before {
		content: '⚠️';
		color: var(--error-color);
		font-size: 1.5em;
		margin-right: 0.5em;
	}

	.button-container {
		display: flex;
		width: 95% !important;
		justify-content: space-around;
		width: 100%; /* Ensuring buttons take full width */
	}

	form {
		display: block;
	}

	@media (max-width: 600px) {
		.button-container {
			flex-direction: column;
		}
	}

	.button-container button {
		margin: 0.2em;
		flex: 1; /* Making buttons equally share the space */
	}

	@media (max-width: 600px) {
		.button-container button {
			flex: none;
			width: 100%;
		}
	}

	.tooltip {
		position: relative;
		font-size: 0.9em;
		cursor: help;
	}

	.tooltip::before {
		content: '⚠️ Not recommended on shared computers';
		position: absolute;
		top: 110%;
		left: 50%;
		transform: translateX(-50%);
		display: none;
		background-color: #aaa;
		color: #000;
		padding: 0.3em 0.6em;
		border-radius: 4px;
		font-size: 1em;
		white-space: nowrap;
	}

	.tooltip:hover::before {
		display: block;
	}

	.spinner {
		border: 4px solid rgba(0, 0, 0, 0.1);
		border-top: 4px solid #3498db;
		border-radius: 50%;
		width: 20px;
		height: 20px;
		animation: spin 1s linear infinite;
		margin-left: 10px;
		display: inline-block;
	}

	@keyframes spin {
		0% {
			transform: rotate(0deg);
		}
		100% {
			transform: rotate(360deg);
		}
	}
</style>
