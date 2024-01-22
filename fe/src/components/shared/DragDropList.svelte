<script lang="ts">
	// Forked from: https://github.com/jwlarocque/svelte-dragdroplist, adapted to Svelte 5  and TS
	import { flip } from 'svelte/animate';

	export let data: any[] = [];
	export let removesItems: boolean = false;

	let ghost: HTMLElement;
	let grabbed: HTMLElement;
	let mouseY: number;
	let offsetY: number;
	let layerY: number;
	let lastTarget: HTMLElement;

	function grab(clientY: number, element: HTMLElement) {
		grabbed = element;
		grabbed.dataset.grabY = clientY.toString();

		ghost.innerHTML = grabbed.innerHTML;

		offsetY = grabbed.getBoundingClientRect().y - clientY;
		drag(clientY);
	}

	function drag(clientY: number) {
		if (grabbed) {
			mouseY = clientY;
			layerY = grabbed.getBoundingClientRect().y - grabbed.offsetTop;
		}
	}

	function touchEnter(touch: Touch) {
		if (grabbed) {
			mouseY = touch.clientY;
			layerY = grabbed.getBoundingClientRect().y - grabbed.offsetTop;
		}
	}

	function dragEnter(target: EventTarget) {
		if (grabbed) {
			if (lastTarget) {
				lastTarget.style.marginBottom = '0.5em';
			}

			if (target instanceof HTMLElement) {
				lastTarget = target;
				target.style.marginBottom = '2.5em';
			}
		}
	}

	function moveDatum(from: number, to: number) {
		let temp = data[from];
		data = [...data.slice(0, from), ...data.slice(from + 1)];
		data = [...data.slice(0, to), temp, ...data.slice(to)];
	}

	function release<T extends MouseEvent | Touch>(ev: T) {
		if (grabbed) {
			// WARN: cast to any here from HTMLElement
			grabbed = null as any;
		}
	}

	function removeDatum(i: number) {
		data = [...data.slice(0, i), ...data.slice(i + 1)];
	}
</script>

<div
	bind:this={ghost}
	id="ghost"
	class={grabbed ? 'item haunting' : 'item'}
	style={'top: ' + (mouseY + offsetY - layerY) + 'px'}
>
	<p />
</div>
<div
	class="list"
	on:mousemove={function (ev) {
		ev.stopPropagation();
		drag(ev.clientY);
	}}
	on:touchmove={function (ev) {
		ev.stopPropagation();
		drag(ev.touches[0].clientY);
	}}
	on:mouseup={function (ev) {
		ev.stopPropagation();
		release(ev);
	}}
	on:touchend={function (ev) {
		ev.stopPropagation();
		release(ev.touches[0]);
	}}
>
	{#each data as datum, i (datum.id ? datum.id : JSON.stringify(datum))}
		<div
			id={grabbed && (datum.id ? datum.id : JSON.stringify(datum)) == grabbed.dataset.id
				? 'grabbed'
				: ''}
			class="item"
			data-index={i}
			data-id={datum.id ? datum.id : JSON.stringify(datum)}
			data-grabY="0"
			on:mousedown={function (ev) {
				grab(window.scrollY, new HTMLElement());
			}}
			on:touchstart={function (ev) {
				grab(ev.touches[0].clientY, new HTMLElement());
			}}
			on:mouseenter={function (ev) {
				ev.stopPropagation();
				if (ev.target) {
					dragEnter(ev.target);
				}
			}}
			on:touchmove={function (ev) {
				ev.stopPropagation();
				ev.preventDefault();
				touchEnter(ev.touches[0]);
			}}
			animate:flip={{ duration: 200 }}
		>
			<div class="buttons">
				<button
					class="up"
					style={'visibility: ' + (i > 0 ? '' : 'hidden') + ';'}
					on:click={function (ev) {
						moveDatum(i, i - 1);
					}}
				>
					<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" width="16px" height="16px"
						><path d="M0 0h24v24H0V0z" fill="none" /><path
							d="M7.41 15.41L12 10.83l4.59 4.58L18 14l-6-6-6 6 1.41 1.41z"
						/></svg
					>
				</button>
				<button
					class="down"
					style={'visibility: ' + (i < data.length - 1 ? '' : 'hidden') + ';'}
					on:click={function (ev) {
						moveDatum(i, i + 1);
					}}
				>
					<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" width="16px" height="16px"
						><path d="M0 0h24v24H0V0z" fill="none" /><path
							d="M7.41 8.59L12 13.17l4.59-4.58L18 10l-6 6-6-6 1.41-1.41z"
						/></svg
					>
				</button>
			</div>

			<div class="content">
				{#if datum.html}
					{@html datum.html}
				{:else if datum.text}
					<p>{datum.text}</p>
				{:else}
					<p>{datum}</p>
				{/if}
			</div>

			<div class="buttons delete">
				{#if removesItems}
					<button
						on:click={function (ev) {
							removeDatum(i);
						}}
					>
						<svg xmlns="http://www.w3.org/2000/svg" height="16" viewBox="0 0 24 24" width="16"
							><path d="M0 0h24v24H0z" fill="none" /><path
								d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z"
							/></svg
						>
					</button>
				{/if}
			</div>
		</div>
	{/each}
</div>

<style>
	main {
		position: relative;
	}

	.list {
		cursor: grab;
		z-index: 5;
		display: flex;
		flex-direction: column;
	}

	.item {
		box-sizing: border-box;
		display: inline-flex;
		width: 100%;
		min-height: 3em;
		margin-bottom: 0.5em;
		background-color: white;
		border: 1px solid rgb(190, 190, 190);
		border-radius: 2px;
		user-select: none;
	}

	.item:last-child {
		margin-bottom: 0;
	}

	.item:not(#grabbed):not(#ghost) {
		z-index: 10;
	}

	.item > * {
		margin: auto;
	}

	.buttons {
		width: 32px;
		min-width: 32px;
		margin: auto 0;
		display: flex;
		flex-direction: column;
	}

	.buttons button {
		cursor: pointer;
		width: 18px;
		height: 18px;
		margin: 0 auto;
		padding: 0;
		border: 1px solid rgba(0, 0, 0, 0);
		background-color: inherit;
	}

	.buttons button:focus {
		border: 1px solid black;
	}

	.delete {
		width: 32px;
	}

	#grabbed {
		opacity: 0;
	}

	#ghost {
		pointer-events: none;
		z-index: -5;
		position: absolute;
		top: 0;
		left: 0;
		opacity: 0;
	}

	#ghost * {
		pointer-events: none;
	}

	#ghost.haunting {
		z-index: 20;
		opacity: 1;
	}
</style>
