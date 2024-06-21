<script lang="ts">
	import { onDestroy } from 'svelte';
	import { browser } from '$app/environment';
	import { _, locale } from 'svelte-i18n';
	import { authStore } from '../../stores/members/auth.ts';
	import PasswordInput from './PasswordInput.svelte';
	import { PasswordMeter } from 'password-meter';

	const tooltipMessage = $_('remember_me_tooltip');
	let isRegistration = false;
	let email_or_username = '';
	if (browser) {
		email_or_username = localStorage.getItem('email_or_username') || '';
	}
	let email = '';
	let nickname = '';
	let sessionTimeMinutes = 30;

	let isEmailAvailable = true;
	let isNickAvailable = true;

	let password = '';
	let showPassword = false;
	let passwordConfirm = '';
	let passwordStrength = '';
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
		const res = await fetch('/api/members/check', 
    {
      method: 'POST',
      headers: headers as HeadersInit,
      body: JSON.stringify(requestPayload)
    });
    if (!res.ok) {
      throw new Error('Network response was not ok');
    }

    const resData = await res.json();

		resData.message === 'available' ? (available = true) : (available = false);
		return available;
	};

const checkNicknameExistApi = async (nickname: string) => {
  const headers = await prepareCSRF();
  const requestPayload = {
    memberName: nickname
  };
  const res = await fetch('/api/members/check', {
    method: 'POST',
    headers: {
      ...headers as HeadersInit,
    },
    body: JSON.stringify(requestPayload)
  });
  const data = await res.json();
  return data.message === 'available';
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
		if (timeoutId) {
			window.clearTimeout(timeoutId);
		}

		timeoutId = window.setTimeout(async () => {
			try {
				strength = new PasswordMeter().getResult(password).score;
				passwordStrength = strength > 136 ? 'Password is strong enough' : `${strength / 2.9}`;
			} catch (error) {
				errorMessage = 'Password is not strong enough or error occurred';
			}
		}, 300);
	};

	const comparePasswords = async (password: string, passwordConfirm: string) => {
		if (password !== passwordConfirm) {
			errorMessage = 'Passwords do not match';
		} else {
			errorMessage = '';
		}
	};

	$: {
		if (isRegistration && password) {
			checkEntropy(password);
		}
	}
	$: tos_url = `/tos/${$locale}`;
	$: privacy_url = `/privacy/${$locale}`;

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

  const headers = new Headers();
  headers.append('Referrer-Policy', 'no-referrer-when-downgrade');
  headers.append('X-CSRF-Token', csrfToken || '');

			localStorage.removeItem('email_or_username');

			// check if passwords match when the registration flow has been triggered
			isRegistration && password !== passwordConfirm
				? ((errorMessage = 'Passwords do not match'), false)
				: passwordStrength !== 'Password is strong enough'
				? ((errorMessage = 'Password is not strong enough'), false)
				: true;

let formData = new FormData();
    formData.append('membername', nickname);
    formData.append('email', email);
    formData.append('password', password);
    formData.append('passwordConfirm', passwordConfirm);
    formData.append('roles', JSON.stringify(['regular']));

 try {
    const response = await fetch('/api/authenticate/register', {
      method: 'POST',
      headers,
      body: formData
    });

    const data = await response.json();

    if (response.ok) {
      authStore.set({
        isAuthenticated: true
      });
      localStorage.removeItem('email_or_username');
      window.location.reload();
      console.info('Registration successful');
      return data.message;
    } else {
      return Promise.reject(data.message);
    }
  } catch (error: any) {
    console.error(error.message);
    return Promise.reject('An error occurred during registration');
  }
    });
	};
	//
	// END OF REGISTRATION
	// START OF LOGIN
	//
	const login = async (event: Event) => {
		event.preventDefault();
		let csrfToken: string | undefined;
		if (browser) {
			csrfToken = document.cookie
				.split('; ')
				.find((row) => row.startsWith('csrf_'))
				?.split('=')[1];
    }
    const headers = new Headers()
  headers.append('Referrer-Policy', 'no-referrer-when-downgrade');
  headers.append('X-CSRF-Token', csrfToken || '');

		localStorage.removeItem('member');

  const formData = new FormData();
  formData.append('membername', email_or_username.includes('@') ? '' : email_or_username);
  formData.append('email', email_or_username.includes('@') ? email_or_username : '');
  formData.append('session_time', sessionTimeMinutes.toString());
  formData.append('password', password);

try {
    const response = await fetch('/api/authenticate/login', {
      method: 'POST',
      headers,
      body: formData
    });

    if (response.ok) {
      authStore.set({
        isAuthenticated: true
      });
      console.debug('authStore updated to ', authStore);
      localStorage.removeItem('email_or_username');
      const responseData = await response.json();

      localStorage.setItem('jwtToken', responseData.token);
      window.location.reload();
      console.info('Login successful');
    } else {
      const data = await response.json();
      console.error(data.message);
      return Promise.reject(data.message);
    }
  } catch (error: any) {
    console.error(error.message);
    return Promise.reject('An error occurred during login');
  }
	};

	onDestroy(() => {
		if (browser) {
			localStorage.removeItem('email_or_username');
		}
	});
</script>

<!-- Form submission handler -->
<form class="auth-form" on:submit|preventDefault={isRegistration ? register : login}>
	<div class="input">
		{#if isRegistration}
			<label class="auth-form-label" for="email">Email:</label>
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
					<p>{$_('email_available')}.</p>
				{:else if !email.match(/(\w+)@(\w+)\.(\w+)/)}
					<span class="spinner" />
				{:else}
					<p class="error-message">
						{$_('email_not_available')}. {$_('try')}
						<a href="https://librate.fediverse.observer/">{$_('another_instance')}</a>
						or <a href="/form/account/recover">{$_('recover_account')}</a>
					</p>
				{/if}
			{/if}
			<label class="auth-form-label" for="nickname">{$_('nickname')}:</label>
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
					<p class="info-message">{$_('nickname_available')}</p>
				{:else}
					<p class="error-message">
						{$_('nickname_not_available')}. {$_('try')}
						<a href="https://librate.fediverse.observer/">{$_('another_instance')}</a>
						{$_('or')}<a href="/form/account/recover">{$_('recover_account')}</a>
					</p>
				{/if}
			{/if}
			<label class="auth-form-label" for="password">{$_('password')}:</label>
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
			<label class="auth-form-label" for="email_or_username">{$_('email_or_username')}:</label>
			<input
				type="text"
				id="email_or_username"
				bind:value={email_or_username}
				required
				class="input"
			/>

			<label class="auth-form-label" for="password">{$_('password')}:</label>
			<PasswordInput
				bind:value={password}
				id="password"
				onInput={async () => void 0}
				{showPassword}
				{toggleObfuscation}
			/>
			<span class="session-timeout-selector">
				<label class="auth-form-label" for="logout-after">
					{$_('logout_after')}
				</label>
				<select class="session-time" bind:value={sessionTimeMinutes}>
					<option value="30">30 {$_('minutes_locative')}</option>
					<option value="60">1 {$_('hour_locative')}</option>
					<option value="120">2 {$_('hours_locative')}</option>
					<option value="360">6 {$_('hours_locative')}</option>
					<option value="720">12 {$_('hours_locative')}</option>
					<option value="1440">1 {$_('day_locative')}</option>
					<option value="10080">1 {$_('week_locative')}</option>
					<option value="2147483647">{$_('never')}</option>
				</select>
			</span>
		{/if}

		{#if isRegistration}
			<label class="auth-form-label" for="password">{$_('confirm')} {$_('password')}:</label>
			<PasswordInput
				id="password"
				bind:value={passwordConfirm}
				onInput={() => comparePasswords(password, passwordConfirm)}
				{showPassword}
				{toggleObfuscation}
			/>
			<!-- Password strength indicator -->
			{#if passwordStrength !== 'Password is strong enough'}
				<p style="padding: 1% 0; display: block;">
          {#if parseInt(passwordStrength) % 10 < 5 }
					{$_('password_strength')}: {passwordStrength} <a
						href="https://www.omnicalculator.com/other/password-entropy"
>{$_('boe_trailing_lt_5')}</a>, {$_('required')}: 50
          {:else}
          {$_('password_strength')}: {passwordStrength} <a
  href="https://www.omnicalculator.com/other/password-entropy"
>{$_('boe_trailing_gte_5')}</a>, {$_('required')}: 50
{/if}
</p>

			{:else}
				<p>
					{$_('password_strength')}: {passwordStrength}
				</p>
			{/if}

			<div class="tos_privacy_ack">
				<input type="checkbox" id="tos_privacy_ack" required />
				<label class="auth-form-label" for="tos_privacy_ack" id="tos_privacy_ack_label">
					{$_('tos_privacy_ack')}
					<a href={privacy_url} target="_blank">{$_('privacy_policy_instrumental')}</a>
					{$_('and')} <a href={tos_url} target="_blank">{$_('tos_instrumental')}</a>
				</label>
			</div>
		{/if}

		{#if errorMessage}
			<p><span class="error-icon" /><span class="error-message">{errorMessage}</span></p>
		{/if}
	</div>
	<!-- End of input container -->
	<div class="button-container">
		{#if !isRegistration}
			<button type="submit" on:click={login}>{$_('sign_in')}</button>
			<button type="button" on:click={startRegistration}>{$_('sign_up')}</button>
		{:else}
			<button type="button" on:click={() => (isRegistration = false)}>{$_('sign_in')}</button>
			<button type="submit" on:click={register}>{$_('sign_up')}</button>
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
		margin: 0.1em 0.5em;
		box-sizing: border-box;
		border: 1px solid #ccc;
		border-radius: 4px;
		width: calc(98% - 0.2em);
		height: 2rem;
		left: 0.2em;
		z-index: -1;
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

	.auth-form {
		display: flex;
		z-index: 1;
		flex-direction: column;
		align-items: center;
	}

	.auth-form-label {
		display: block;
		padding: 1% 0.2%;
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
	/* removed unused selectors for tooltip, shall it be reimplemented,
  see the commit from Feb 01 2024 */

	.tos_privacy_ack {
		display: inline-flex;
		font-size: 80%;
		float: inline-start;
		align-items: center;
	}

	#tos_privacy_ack_label {
		margin-left: 2%;
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
