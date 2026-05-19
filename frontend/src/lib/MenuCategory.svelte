<script lang="ts">
  import MenuCard from './MenuCard.svelte';
  import type { Dish } from './constants';

  let {
    title,
    subtitle,
    dishes,
    isExpanded,
    toggleExpand,
    getQty,
    onAdd,
    onRemove
  }: {
    title: string;
    subtitle: string;
    dishes: Dish[];
    isExpanded: boolean;
    toggleExpand: () => void;
    getQty: (id: string) => number;
    onAdd: (id: string) => void;
    onRemove: (id: string) => void;
  } = $props();
</script>

<div class="space-y-16">
  <div class="flex items-end justify-between border-b border-white/5 pb-8">
    <div class="space-y-1">
      <h3 class="text-4xl font-display font-light uppercase tracking-tight">{title}</h3>
    </div>
  </div>

  <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
    {#each dishes as item (item.id)}
      <MenuCard 
        {item} 
        qty={getQty(item.id)} 
        onAdd={() => onAdd(item.id)}
        onRemove={() => onRemove(item.id)}
      />
    {/each}
  </div>

</div>
