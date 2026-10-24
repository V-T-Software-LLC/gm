<script>
  let count = $state(0)
</script>

<button
  class="rounded bg-blue-600 px-4 py-2 text-white hover:bg-blue-700"
  onclick={() => count++}
>
  count is {count}
</button>
