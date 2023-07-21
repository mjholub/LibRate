<script lang="ts">
	import { onMount } from 'svelte';
	import { browser } from '$app/environment';
	import { authStore } from '../../stores/members/auth.ts';
	import PasswordInput from './PasswordInput.svelte';
	import type { Member } from '../../types/member.ts';

	let isRegistration = false;
	let member: Member;
	let isAuthenticated = authStore.authenticate();
	let email_or_username = '';
	if (browser) {
		email_or_username = localStorage.getItem('email_or_username') || '';
	}
	let email = '';
	let nickname = '';
	let password = '';
	let showPassword = false;
	let passwordConfirm = '';
	let passwordStrength = '' as string; // it is based on the message from the backend, not the entropy score
	let errorMessage = '';

	const toggleObfuscation = () => {
		showPassword = !showPassword;
	};

	// helper function to check password strength
	let timeoutId: any;
	const checkEntropy = async (password: string) => {
		// if just logging in, don't check the entropy
		if (!isRegistration) return;

		clearTimeout(timeoutId);
		timeoutId = setTimeout(async () => {
			const response = await fetch(`/api/password-entropy`, {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json'
				},
				body: JSON.stringify({ password })
			});
			const data = await response.json();
			passwordStrength = data.message;
		}, 300);
	};

	const entropyDummy = async (password: string) => {
		Promise.resolve(password);
	};

	$: isRegistration && password && checkEntropy(password);

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

		isRegistration && password !== passwordConfirm
			? ((errorMessage = 'Passwords do not match'), false)
			: passwordStrength !== 'Password is strong enough'
			? ((errorMessage = 'Password is not strong enough'), false)
			: true;

		const response = await fetch('/api/members/register', {
			method: 'POST',
			headers: {
				'Content-Type': 'application/json'
			},
			body: JSON.stringify({
				MemberName: nickname,
				Email: email,
				Password: password,
				PasswordConfirm: passwordConfirm,
				Roles: ['regular']
			})
		});

		const data = await response.json();

		if (browser) {
			response.ok
				? (localStorage.setItem('token', data.token),
				  localStorage.setItem('email_or_username', ''),
				  await authStore.authenticate(),
				  authStore.set(data.member),
				  authStore.getMember(data.member.id),
				  (window.location.href = '/'))
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
				MemberName: email_or_username.includes('@') ? '' : email_or_username,
				Email: email_or_username.includes('@') ? email_or_username : '',
				Password: password
			})
		});

		const data = await response.json();
		console.debug(data);

		response.ok
			? (localStorage.setItem('token', data.token),
			  localStorage.setItem('email_or_username', ''),
			  authStore.set(data.member),
			  await authStore.authenticate(),
			  authStore.getMember(data.member.id),
			  (window.location.href = '/'),
			  console.info('Login successful'))
			: (errorMessage = data.message);
		console.error(data.message);
	};
</script>

<!-- Form submission handler -->
<form on:submit|preventDefault={isRegistration ? register : login}>
	{#if !isRegistration}
		<label for="email_or_username">Email or Username:</label>
		<input
			type="text"
			id="email_or_username"
			bind:value={email_or_username}
			required
			aria-label="Email or Username"
		/>

		<PasswordInput
			bind:value={password}
			id="password"
			onInput={entropyDummy}
			{showPassword}
			{toggleObfuscation}
		/>
	{:else}
		<!-- Registration form -->
		<label for="email">Email:</label>
		<input id="email" bind:value={email} type="email" required aria-label="Email" />

		<label for="nickname">Nickname:</label>
		<input id="nickname" bind:value={nickname} required aria-label="Nickname" />

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

	{#if !isRegistration}
		<button type="submit" on:click={login}>Sign In</button>
		<button type="button" on:click={startRegistration}>Sign Up</button>
	{:else}
		<button type="submit">Sign Up</button>
		<button type="button" on:click={() => (isRegistration = false)}>Sign In</button>
	{/if}
</form>

<style>
	input,
	button {
		font-family: inherit;
		font-size: inherit;
		padding: 0.4em;
		margin: 0 0 0.5em 0;
		box-sizing: border-box;
		border: 1px solid #ccc;
		border-radius: 4px;
	}

	.error-message {
		color: red;
		font-weight: bold;
	}
</style>
