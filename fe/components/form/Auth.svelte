<script>
  import { onMount } from "svelte";

  let isRegistration = false;
  let email_or_username = "";
  let password = "";
  let showPassword = false;
  let confirmPassword = "";
  let passwordStrength = 0;
  let errorMessage = "";

  const checkEntropy = async (password) => {
    const response = await fetch(`/api/password-entropy`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ password }),
    });
    const data = await response.json();
    passwordStrength = data.message;
  };

  $: password && checkEntropy(password);

  const register = async (event) => {
    event.preventDefault();

    if (isRegistration && password !== confirmPassword) {
      errorMessage = "Passwords do not match";
      return;
    }

    if (passwordStrength !== "Password is strong enough") {
      errorMessage = "Password is not strong enough";
      return;
    }

    const response = await fetch("/api/register", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        MemberName: email_or_username.includes("@") ? "" : email_or_username,
        Email: email_or_username.includes("@") ? email_or_username : "",
        Password: password,
      }),
    });

    const data = await response.json();

    if (response.ok) {
      const { token } = data;
      localStorage.setItem("token", token);
      window.location = "/";
    } else {
      errorMessage = data.message;
    }
  };

  const login = async (event) => {
    event.preventDefault();

    const response = await fetch("/api/login", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        MemberName: email_or_username.includes("@") ? "" : email_or_username,
        Email: email_or_username.includes("@") ? email_or_username : "",
        Password: password,
      }),
    });

    const data = await response.json();

    if (response.ok) {
      const { token } = data;
      localStorage.setItem("token", token);
      window.location = "/";
    } else {
      errorMessage = data.message;
    }
  };
</script>

<form on:submit|preventDefault={isRegistration ? register : login}>
  <label for="email_or_username">Email or username:</label>
  <input id="email_or_username" bind:value={email_or_username} required />

  <label for="password">Password:</label>
  <input
    id="password"
    bind:value={password}
    type="password"
    required
    class={!showPassword ? "" : "hidden"}
  />

  <input
    id="textPassword"
    bind:value={password}
    type="text"
    required
    class={showPassword ? "" : "hidden"}
  />

  <button
    class="show-password"
    type="button"
    on:click={() => (showPassword = !showPassword)}
  >
    {showPassword ? "Hide" : "Show"} password
  </button>

  {#if isRegistration}
    <label for="confirmPassword">Confirm Password:</label>
    <input
      id="confirmPassword"
      bind:value={confirmPassword}
      type="password"
      required
    />
  {/if}

  <p>Password strength: {passwordStrength} bits of entropy, required: 60</p>

  {#if errorMessage}
    <p style="color: red;">{errorMessage}</p>
  {/if}

  {#if !isRegistration}
    <button type="submit">Sign In</button>
    <button type="button" on:click={() => (isRegistration = true)}
      >Sign Up</button
    >
  {:else}
    <button type="submit">Sign Up</button>
    <button type="button" on:click={() => (isRegistration = false)}
      >Sign In</button
    >
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

  .show-password {
    margin-left: 0.5em;
    position: relative;
    border: none;
  }
</style>
