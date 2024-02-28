<script lang="ts">
import * as Card from "$components/ui/card"
import Label from "$components/ui/label/label.svelte";
import Footer from '$components/footer/footer.svelte';
import Header from '$components/header/Header.svelte';
import type { SearchResponse } from '$lib/types/search.ts';
import { _, locale } from 'svelte-i18n'; 
import { memberStore } from "$stores/members/getInfo";
	import type { Member } from "$lib/types/member";
	import MemberCard from "$components/member/MemberCard.svelte";
	import { stringify } from "uuid";

export let results: SearchResponse;
let theme = "default";

const getLocalizedDescription = (descriptions: any[]): any => {
  return descriptions.find((desc: Record<string, any>) => desc.language === $locale).description
};

const jwtToken = localStorage.getItem('jwtToken') || '';
const getMemberByWebfinger = async (webfinger: string): Promise<Member> => {
  if (jwtToken === '') {
    throw new Error('Unauthorized'); 
  }
  // TODO: modify when federation is added
  const memberName = webfinger.split("@")[0];
  const memberData = await memberStore.getMember(jwtToken, memberName);
  return memberData;
}
</script>
{#if results.totalHits === 0}
<div class="space-y-1">
  <h3>${$_('no_results_found')}</h3>
</div>
{:else}
<div class="space-y-1">
  <h3>{results.totalHits}{$_('results')}</h3>
  <h6>{$_('processing_time')}: {results.processingTime} ms</h6>
</div>
{#each results.categories as category}
<h3>{category.toString()}</h3>
<div class="container">  
    <!-- TODO: server side localization -->
   <Card.Root>
    {#if category === "genres"}
    {#each results.data as [key, value]}
      <Card.Header>
      <Card.Title><a href="{value.url}">{value.name}</a>
      </Card.Title>
      <Card.Description>
        {#each value.kinds as kind}
        {$_(kind)}
        {/each}
      </Card.Description>
</Card.Header>
<Card.Content>
  {#each value.descriptions as descs}
    {getLocalizedDescription(descs)}
  {/each}
  </Card.Content>
    {/each}
    {:else if category === "members"}
      {#each results.data as [key, value]}
        {#await getMemberByWebfinger(value.webfinger)}
        <span class="spinner"/>
        {:then member}
          <MemberCard {member} />
        {:catch error}
          <p class="error">{error}</p>
        {/await}
      {/each}
    {/if}
</Card.Root>
   </div>     
   <hr />
  {/each}  
  {/if}
<Footer />
<style>
  .container {
    margin: 0 20%;
  }
  .category {
    display: flex;
    flex-direction: row;
    overflow-x: auto;
  }
  .item {
    flex: 0 0 auto;
    margin-right: 1em;
  }
</style>
