<script lang="ts">
	import axios from 'axios';
	import { onMount } from 'svelte';
	import { browser } from '$app/environment';
	import { authStore } from '../../stores/members/auth.ts';
	import PasswordInput from './PasswordInput.svelte';
	import type { AuthStoreState } from '$stores/members/auth.ts';

	let tooltipMessage = 'This feature is not implemented yet';
	let isRegistration = false;
	let email_or_username = '';
	if (browser) {
		email_or_username = localStorage.getItem('email_or_username') || '';
	}
	let authToken = null;
	let email = '';
	let nickname = '';
	let password = '';
	let showPassword = false;
	let passwordConfirm = '';
	let passwordStrength = '' as string; // it is based on the message from the backend, not the entropy score
	let errorMessage = '';
	let authState: AuthStoreState = $authStore;

	const toggleObfuscation = () => {
		showPassword = !showPassword;
	};

	// helper function to check password strength
	let timeoutId: number | undefined;
	const checkEntropy = async (password: string) => {
		// if just logging in, don't check the entropy
		if (!isRegistration) return;

		if (timeoutId) {
			window.clearTimeout(timeoutId);
		}

		timeoutId = window.setTimeout(async () => {
			try {
				const response = await fetch(`/api/password-entropy`, {
					method: 'POST',
					headers: {
						'Content-Type': 'application/json'
					},
					body: JSON.stringify({ password })
				});

				const data = await response.json();
				passwordStrength = data.message;
			} catch (error) {
				console.error(error);
			}
		}, 300);
	};
	$: {
		if (isRegistration && password) {
			checkEntropy(password);
		}
	}

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
		event.preventDefault();

		// check if passwords match when the registration flow has been triggered
		isRegistration && password !== passwordConfirm
			? ((errorMessage = 'Passwords do not match'), false)
			: passwordStrength !== 'Password is strong enough'
			? ((errorMessage = 'Password is not strong enough'), false)
			: true;

		const response = await axios.post('/api/members/register', {
			membername: nickname,
			email: email,
			password: password,
			passwordConfirm: passwordConfirm,
			roles: ['regular']
		});

		const data = response.data;

		const member = await authStore.getMember(data.member_id);

		if (browser) {
			const registrationSuccessful = response.status == 200 && data.member_id !== 0;
			registrationSuccessful
				? ((authToken = data.token),
				  await authStore.authenticate(),
				  localStorage.setItem('email_or_username', ''),
				  authStore.set(data.member),
				  (authState.id = member.id),
				  (window.location.href = '/'),
				  console.debug('Registration successful'))
				: (errorMessage = data.message);
		}
	};

	const login = async (event: Event) => {
		event.preventDefault();

		const response = await fetch('/api/members/login', {
			method: 'POST',
			headers: {
				'Content-Type': 'application/json'
			},
			body: JSON.stringify({
				membername: email_or_username.includes('@') ? '' : email_or_username,
				email: email_or_username.includes('@') ? email_or_username : '',
				password: password
			})
		});

		const data = await response.json();
		const member = await authStore.getMember(data.member_id);

		if (browser) {
			response.ok
				? (localStorage.setItem('token', data.token),
				  await authStore.authenticate(),
				  authStore.set({
						...member, // Include existing member properties
						id: data.member_id,
						isAuthenticated: true
				  }),
				  localStorage.setItem('member', JSON.stringify(member)),
				  localStorage.setItem('email_or_username', ''),
				  (authState.id = member.id),
				  (window.location.href = '/'),
				  console.info('Login successful'))
				: (errorMessage = data.message);
			console.error(data.message);
		}
	};
</script>

<!-- Form submission handler -->
<form on:submit|preventDefault={isRegistration ? register : login}>
	<div class="input">
		{#if !isRegistration}
			<label for="email_or_username">Email or Username:</label>
			<input
				type="text"
				id="email_or_username"
				bind:value={email_or_username}
				required
				class="input"
				aria-label="Email or Username"
			/>

			<PasswordInput
				bind:value={password}
				id="password"
				onInput={() => checkEntropy(password)}
				{showPassword}
				{toggleObfuscation}
			/>
			<label for="rememberMe"
				>Remember me<span class="tooltip" aria-label={tooltipMessage}> *</span></label
			>
			<input type="checkbox" id="rememberMe" name="rememberMe" value="rememberMe" />
		{:else}
			<!-- Registration form -->
			<label for="email">Email:</label>
			<input id="email" bind:value={email} type="email" class="input" required aria-label="Email" />

			<label for="nickname">Nickname:</label>
			<input id="nickname" bind:value={nickname} required aria-label="Nickname" class="input" />

			<PasswordInput
				bind:value={password}
				id="password"
				onInput={() => checkEntropy(password)}
				{showPassword}
				{toggleObfuscation}
			/>
		{/if}

		{#if isRegistration}
			<label for="passwordConfirm">Confirm Password:</label>
			<input
				id="passwordConfirm"
				class="input"
				bind:value={passwordConfirm}
				type="password"
				required
				aria-label="Confirm Password"
				on:input={() => checkEntropy(passwordConfirm)}
			/>
			<!-- Password strength indicator -->
			{#if passwordStrength !== 'Password is strong enough'}
				<p>
					Password strength: {passwordStrength} bits of (<a
						href="https://www.omnicalculator.com/other/password-entropy">entropy</a
					>), required: 50
				</p>
			{:else}
				<p>Password strength: {passwordStrength}</p>
			{/if}
		{/if}

		{#if errorMessage}
			<p class="error-message">{errorMessage}</p>
		{/if}
	</div>
	<!-- End of input container -->
	<div class="button-container">
		{#if !isRegistration}
			<button type="submit" on:click={login}>Sign In</button>
			<button type="button" on:click={startRegistration}>Sign Up</button>
		{:else}
			<button type="submit">Sign Up</button>
			<button type="button" on:click={() => (isRegistration = false)}>Sign In</button>
		{/if}
	</div>
</form>

<style>
	.input {
		font-family: inherit;
		font-size: inherit;
		padding: 0.4em;
		margin: 0 0 0.5em 0;
		box-sizing: border-box;
		border: 1px solid #ccc;
		border-radius: 4px;
		width: 100%; /* Ensuring inputs take full width */
	}

	.error-message {
		color: red;
		font-weight: bold;
	}

	.button-container {
		display: flex;
		justify-content: space-around;
		width: 100%; /* Ensuring buttons take full width */
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
		content: '⚠️ This feature is not implemented yet';
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
</style>
