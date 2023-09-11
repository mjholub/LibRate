<script lang="ts">
	import { getContext } from 'svelte';

	const params: any = getContext('route');
	let user = params.user;
</script>

<svelte:head>
	<title>LibRate - {user}'s Profile</title>
</svelte:head>
<slot />
