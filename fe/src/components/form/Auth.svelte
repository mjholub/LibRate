<script>
  import { onMount } from "svelte";

  let isRegistration = false;
  let email_or_username = localStorage.getItem("email_or_username") || "";
  let email = "";
  let nickname = "";
  let password = "";
  let showPassword = false;
  let passwordConfirm = "";
  let passwordStrength = 0;
  let errorMessage = "";

  const toggleObfuscation = () => {
    showPassword = !showPassword;
  };

  // helper function to check password strength
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

  // helper function to trigger moving either email or nickname to a dedicated field
  const startRegistration = () => {
    isRegistration = true;
    email_or_username.includes("@")
      ? ((email = email_or_username),
        localStorage.setItem("email_or_username", ""))
      : ((nickname = email_or_username),
        localStorage.setItem("email_or_username", ""));
  };

  const register = async (event) => {
    event.preventDefault();

    isRegistration && password !== passwordConfirm
      ? ((errorMessage = "Passwords do not match"), false)
      : passwordStrength !== "Password is strong enough"
      ? ((errorMessage = "Password is not strong enough"), false)
      : true;

    const response = await fetch("/api/register", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        MemberName: nickname,
        Email: email,
        Password: password,
        PasswordConfirm: passwordConfirm,
      }),
    });

    const data = await response.json();

    response.ok
      ? (localStorage.setItem("token", data.token),
        localStorage.setItem("email_or_username", ""),
        (window.location = "/"))
      : (errorMessage = data.message);
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
  {#if !isRegistration}
    <label for="email_or_username">Email or Username:</label>
    <input
      type="text"
      id="email_or_username"
      bind:value={email_or_username}
      required
      aria-label="Email or Username"
    />

    <label for="password">Password:</label>
    <div class="password-container">
      <input
        id="password"
        class={!showPassword ? "" : "hidden"}
        bind:value={password}
        type="password"
        autocomplete="new-password"
        required
        aria-label="Password"
      />
      <input
        id="textPassword"
        class={showPassword ? "" : "hidden"}
        bind:value={password}
        type="text"
        autocomplete="new-password"
        required
        aria-label="Password"
      />
      <button
        class="toggle-btn"
        type="button"
        on:click|preventDefault={toggleObfuscation}
        aria-label="Toggle password visibility"
      >
        <span class="material-icons">
          {showPassword ? "visibility_off" : "visibility"}
        </span>
      </button>
    </div>
  {:else}
    <label for="email">Email:</label>
    <input
      id="email"
      bind:value={email}
      type="email"
      required
      aria-label="Email"
    />

    <label for="nickname">Nickname:</label>
    <input id="nickname" bind:value={nickname} required aria-label="Nickname" />

    <label for="passwordConfirm">Confirm Password:</label>
    <input
      id="passwordConfirm"
      bind:value={passwordConfirm}
      type="password"
      required
      aria-label="Confirm Password"
    />

    <p>Password strength: {passwordStrength} bits of entropy, required: 60</p>
  {/if}

  {#if errorMessage}
    <p class="error-message">{errorMessage}</p>
  {/if}

  {#if !isRegistration}
    <button type="submit">Sign In</button>
    <button type="button" on:click={startRegistration}>Sign Up</button>
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
    overflow: hidden;
    display: flex;
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

  .error-message {
    color: red;
    font-weight: bold;
  }
</style>