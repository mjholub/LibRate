<script>
  import { onMount } from "svelte";

  let email_or_username = "";
  let password = "";
  let confirmPassword = "";
  let passwordStrength = "";
  let errorMessage = "";

  const checkEntropy = async (password) => {
    const response = await fetch(`/api/password-entropy?password=${password}`);
    const data = await response.json();
    passwordStrength = data.entropy;
  };

  // reactive statement to watch `password`
  $: if (password) {
    checkEntropy(password)
      .then((entropy) => {
        passwordStrength = entropy;
      })
      .catch((error) => {
        console.error("Error:", error);
      });
  }

  const register = async () => {
    if (password !== confirmPassword) {
      errorMessage = "Passwords do not match";
      return;
    }

    if (passwordStrength < 60) {
      errorMessage = "Password is not strong enough";
      return;
    }

    const response = await fetch("/api/register", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        MemberName: email_or_username,
        Email: email_or_username,
        Password: password,
        PasswordConfirm: confirmPassword,
      }),
    });

    if (response.ok) {
      // auto login the new user and render a confirmation prompt top bar
      const { token } = await response.json();
      if (token) {
        localStorage.setItem("token", token);
        window.location = "/";
      }
    } else {
      const { message } = await response.json();
      errorMessage = message;
    }
    // end of token validation branch
  };
  // end of response.ok branch

  $: password && checkEntropy();
</script>

<form on:submit|preventDefault={register}>
  <label for="email">Email or username:</label>
  <input id="email_or_username" bind:value={email_or_username} required />

  <label for="password">Password:</label>
  <input id="password" bind:value={password} type="password" required />

  <label for="confirmPassword">Confirm Password:</label>
  <input
    id="confirmPassword"
    bind:value={confirmPassword}
    type="password"
    required
  />

  <p>Password strength: {passwordStrength} bits</p>

  {#if errorMessage}
    <p style="color: red;">{errorMessage}</p>
  {/if}

  <button type="submit">Sign Up</button>
</form>

<style>
  input,
  button {
    font-family: inherit;
    font-size: inherit;
    -webkit-padding: 0.4em 0;
    padding: 0.4em;
    margin: 0 0 0.5em 0;
    box-sizing: border-box;
    border: 1px solid #ccc;
    border-radius: 4px;
  }
</style>
