import type { SvelteComponent } from "svelte"

export type TabItem = {
index: number;
label: string;
componentName: any;
events: Record<string, (event: CustomEvent<any>) => void>;
}