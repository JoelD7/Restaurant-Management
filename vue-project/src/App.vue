<template>
  <v-app class="app">
    <v-main>
      <router-view />
    </v-main>
  </v-app>
</template>

<script lang="ts">
import Vue from "vue";

export default Vue.extend({
  name: "App",

  data: () => ({
    //
  }),
});
</script>

<style>
@import url("https://fonts.googleapis.com/css2?family=Poppins&display=swap");

.app {
  font-family: "Poppins", sans-serif;
}
</style>