<script lang="ts">
  import { Plus, Minus } from '@lucide/svelte';
  import type { Dish } from './constants';

  let { item, qty, onAdd, onRemove, onCardClick }: {
    item: Dish;
    qty: number;
    onAdd: () => void;
    onRemove: () => void;
    onCardClick?: () => void;
  } = $props();
</script>

<div
  onclick={() => onCardClick?.()}
  class="bg-white/[0.01] border border-white/5 p-8 group transition-all duration-700 hover:border-white/20 hover:-translate-y-1 relative overflow-hidden {onCardClick ? 'cursor-pointer' : ''}"
>
  <div class="relative mb-8 aspect-square overflow-hidden bg-[#0a0a0a]">
    <img
      src={item.image}
      alt={item.name}
      class="w-full h-full object-cover grayscale opacity-60 group-hover:grayscale-0 group-hover:opacity-100 group-hover:scale-105 transition-all duration-1000 ease-out"
      referrerpolicy="no-referrer"
    />
    <div class="absolute top-4 right-4 opacity-0 group-hover:opacity-100 transition-opacity">
      <div class="w-8 h-8 rounded-full bg-white/10 backdrop-blur-md flex items-center justify-center">
        <Plus class="w-3 h-3 text-white" />
      </div>
    </div>
  </div>

  <div class="space-y-3 relative z-10">
    <div class="space-y-2">
      <h4 class="font-medium text-[10px] leading-relaxed line-clamp-2 uppercase tracking-widest text-white/60 group-hover:text-white transition-colors">
        {item.name}
      </h4>
      {#if item.description}
        <p class="text-[11px] text-white/30 leading-relaxed font-light line-clamp-2 group-hover:text-white/50 transition-colors">
          {item.description}
        </p>
      {/if}
    </div>

    <!-- Calories / weight hint -->
    {#if item.calories || item.weight}
      <div class="flex items-center gap-3 text-[9px] font-mono text-white/20">
        {#if item.calories}<span>{item.calories} ккал</span>{/if}
        {#if item.calories && item.weight}<span class="text-white/10">·</span>{/if}
        {#if item.weight}<span>{item.weight} г</span>{/if}
      </div>
    {/if}

    <div class="flex items-center justify-between pt-4 border-t border-white/5">
      <span class="text-xl font-mono tracking-tighter text-white/80">{item.price}.00 ₽</span>

      {#if qty > 0}
        <div class="flex items-center gap-4 bg-white/5 border border-white/10 p-1 px-4 rounded-sm">
          <button
            onclick={(e) => { e.stopPropagation(); onRemove(); }}
            class="text-white/20 hover:text-white transition-colors cursor-pointer"
          >
            <Minus class="w-3 h-3" />
          </button>
          <span class="font-mono text-xs text-white">{qty}</span>
          <button
            onclick={(e) => { e.stopPropagation(); onAdd(); }}
            class="text-brand-red hover:text-red-400 transition-colors cursor-pointer"
          >
            <Plus class="w-3 h-3" />
          </button>
        </div>
      {:else}
        <button
          onclick={(e) => { e.stopPropagation(); onAdd(); }}
          class="text-[10px] font-bold uppercase tracking-[0.2em] text-white/40 hover:text-brand-red transition-all cursor-pointer"
        >
          Добавить
        </button>
      {/if}
    </div>
  </div>
</div>
