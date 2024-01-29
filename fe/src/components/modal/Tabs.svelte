<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { _ } from 'svelte-i18n';
  import type {TabItem} from '$lib/types/tabs';
  export let items: TabItem[] = [];
  export let activeTabIndex = 0 
  const handleClick = (tabIndex: number) => () => (activeTabIndex = tabIndex)
  
  const handleKeyDown = (event: KeyboardEvent) => {
    if (event.key === 'Shift') {
      event.preventDefault();
    } else if (event.key === 'ArrowRight' || event.key === 'ArrowLeft') {
      const direction = event.key === 'ArrowRight' ? 1 : -1;
      const newIndex = (activeTabIndex + direction + items.length) % items.length;
      activeTabIndex = newIndex;
    }
  }
  
  onMount(() => {
    window.addEventListener('keydown', handleKeyDown);
  })

  onDestroy(() => {
    window.removeEventListener('keydown', handleKeyDown);
    })

  </script>
  
  <ul>
    {#each items as item}
      <li class={activeTabIndex === item.index ? 'active' : ''} />
      <!-- svelte-ignore a11y-click-events-have-key-events -->
      <!-- handled by the onMount event listener -->
      <span on:click={handleClick(item.index)}
      tabindex="-1"
      role="tabpanel"
      >{item.label}</span>
    {/each}
  </ul>
  <p class="info-msg">{$_('key-combination-infomsg-tabs')}</p>
  {#each items as item}
  {#if activeTabIndex == item.index}
  <div class="box">
    <this={item.component} />
  </div>
  {/if}
  {/each}
  
  <style>
    .box {
      margin-bottom: 10px;
      padding: 40px;
      border: 1px solid #dee2e6;
      border-radius: 0 0 .5rem .5rem;
      border-top: 0;
    }
    ul {
      display: flex;
      flex-wrap: wrap;
      padding-left: 0;
      margin-bottom: 0;
      list-style: none;
      border-bottom: 1px solid #dee2e6;
    }
    li {
      margin-bottom: -1px;
    }
  
    span {
      border: 1px solid transparent;
      border-top-left-radius: 0.25rem;
      border-top-right-radius: 0.25rem;
      display: block;
      padding: 0.5rem 1rem;
      cursor: pointer;
    }
  
    span:hover {
      border-color: #e9ecef #e9ecef #dee2e6;
    }
  
    li.active > span {
      color: #495057;
      background-color: #fff;
      border-color: #dee2e6 #dee2e6 #fff;
    }
  </style>