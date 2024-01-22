<script lang="ts">
	import type { Book } from '$lib/types/books';
	import { Label, Input, Collapse } from '@sveltestrap/sveltestrap';
	import { genreStore } from '$stores/media/genre';

	let genreNames: string[] = [];

	let book: Book = {
		media_id: '',
		title: '',
		authors: null,
		publisher: '',
		publication_date: new Date(),
		genres: [],
		keywords: [],
		languages: [],
		pages: 0,
		isbn: '',
		asin: '',
		cover: '',
		summary: ''
	};

	const fetchGenreNames = async (message: string) => {
		console.debug(`genre names are ${message}, getting genre names from API endpoint`);
		const genresResponse = await genreStore.getGenreNames('book', false);
		genreNames.push(...genresResponse);
		genreNames = [...genreNames];
		localStorage.setItem('genreNames', JSON.stringify(genreNames));
		localStorage.setItem(`genreNames_timestamp`, new Date().getTime().toString());
	};

	const submitBook = async () => {
		alert('Warned you, this is just a mockup');
	};
</script>

<form class="sub-form" on:submit|preventDefault={submitBook}>
	<p class="warning">Book submission not working yet, this is just a mockup.</p>
	<Label for="title">Title</Label>
	<Input type="text" id="title" bind:value={book.title} />
	<!-- collapsible section for adding cast -->
	<Collapse
		>click here to start adding authors
		<Label for="authors">Authors</Label>
	</Collapse>
	<Label for="publisher">Publisher</Label>
	<Input type="text" id="publisher" bind:value={book.publisher} />
	<Input type="date" id="publication_date" bind:value={book.publication_date} />
	<Label for="pages">Pages</Label>
	<Input type="number" id="pages" bind:value={book.pages} />
	<Label for="isbn">ISBN</Label>
	<Input type="text" id="isbn" bind:value={book.isbn} />
	<Label for="asin">ASIN</Label>
	<Input type="text" id="asin" bind:value={book.asin} />
	<Label for="summary">Summary</Label>
	<Input type="textarea" id="summary" bind:value={book.summary} />
	<button type="submit">Submit</button>
</form>

<style>
	.sub_form {
		display: block;
	}
	label {
		display: block;
		margin-top: 1rem;
	}
	input {
		margin-top: 0.4rem;
		margin-bottom: 0.2rem;
	}

	.warning {
		color: red;
		font-weight: bold;
		font-size: 1.5em;
	}
</style>
