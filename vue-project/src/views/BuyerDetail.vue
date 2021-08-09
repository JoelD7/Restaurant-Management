<template>
  <div>
    {{ $route.params.id }}
  </div>
</template>

<script lang="ts">
import Vue from "vue";

export default Vue.extend({
  name: "BuyerDetail",
});
</script>

<style scoped>
</style>