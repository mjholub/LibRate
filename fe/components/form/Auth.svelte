<script>
  import { onMount } from "svelte";

  let isRegistration = false;
  let email_or_username = "";
  let password = "";
  let showPassword = false;
  let passwordConfirm = "";
  let passwordStrength = 0;
  let errorMessage = "";

  const toggleObfuscation = () => {
    showPassword = !showPassword;
  };

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

    if (isRegistration && password !== passwordConfirm) {
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
        PasswordConfirm: passwordConfirm,
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

  <div class="password-container">
    <label for="password">Password:</label>

    <input
      id="password"
      class={!showPassword ? "" : "hidden"}
      bind:value={password}
      type="password"
      autocomplete="new-password"
      required
    />

    <!-- use the autocomplete property to prevent the browser from filling in the password -->
    <input
      id="textPassword"
      class={showPassword ? "" : "hidden"}
      bind:value={password}
      type="text"
      autocomplete="new-password"
      required
    />

    <button
      class="toggle-btn"
      type="button"
      on:click|preventDefault={toggleObfuscation}
    >
      <span class="material-icons">
        {showPassword ? "visibility_off" : "visibility"}
      </span>
    </button>
  </div>

  {#if isRegistration}
    <label for="passwordConfirm">Confirm Password:</label>
    <input
      id="passwordConfirm"
      bind:value={passwordConfirm}
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

  .password-container {
    position: relative;
    display: inline-block;
  }

  .hidden {
    display: none;
  }

  .toggle-btn {
    position: absolute;
    right: 10px;
    top: 50%;
    transform: translateY(-50%);
    background: transparent;
    border: none;
    cursor: pointer;
  }

  .material-icons {
    font-family: "Material Icons";
    font-weight: normal;
    font-style: normal;
    font-size: 20px; /* Preferred icon size */
    display: inline-block;
    line-height: 1;
    text-transform: none;
    letter-spacing: normal;
    word-wrap: normal;
    white-space: nowrap;
    direction: ltr;
    -webkit-font-smoothing: antialiased;
    text-rendering: optimizeLegibility;
    -moz-osx-font-smoothing: grayscale;
    font-feature-settings: "liga";
  }
</style>
