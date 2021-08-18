<template>
  <v-container class="main-container">
    <!-- Navbar -->
    <div class="navbar">
      <v-btn text href="/" class="navbar-btn">Inicio</v-btn>
    </div>

    <v-row align="center" no-gutters style="height: 100vh">
      <v-col>
        <v-row justify="center" align="center" no-gutters>
          <h1 :style="{ color: Colors.BLUE }">No encontrado</h1>
        </v-row>

        <v-row justify="center" align="center" no-gutters>
          <p>
            La página solicitada no existe. Haga click en
            <b :style="{ color: Colors.GREEN }">Inicio</b> para ir a la página
            principal.
          </p>
        </v-row>

        <v-row justify="center" align="center" no-gutters>
          <v-btn
            class="btn"
            @click="goToHome()"
            :style="{ 'background-color': Colors.GREEN }"
          >
            Inicio
          </v-btn>
        </v-row>
      </v-col>
    </v-row>
  </v-container>
</template>

<script lang="ts">
import Vue from "vue";
import { Colors } from "../assets/colors";
export default Vue.extend({
  data() {
    return { Colors };
  },

  methods: {
    goToHome() {
      this.$router.push({ path: "/" });
    },
  },
});
</script>
    
<style scoped>
.btn {
  text-transform: capitalize !important;
  color: white !important;
}
.main-container {
  font-family: "Poppins", sans-serif;
  height: 100vh;
  width: 100vw;
}

.navbar {
  display: flex;
  margin-top: 12px;
}

.navbar-btn {
  color: #004e88 !important;
  margin: 0px 5px !important;
  text-transform: capitalize !important;
}
</style>