<template>
  <v-dialog persistent width="700" content-class="error-dialog" v-model="open">
    <v-card class="card">
      <v-card-title class="card-title">
        <font-awesome-icon
          size="lg"
          style="margin-right: 10px"
          :color="Colors.RED_ERROR"
          :icon="['fas', 'exclamation-triangle']"
        />
        Un error ha ocurrido
      </v-card-title>

      <v-card-subtitle style="margin-top: 5px">
        Ha ocurrido un error. Refresque la página e inténtelo nuevamente.
      </v-card-subtitle>

      <v-divider></v-divider>

      <v-card-text style="color: black; margin-top: 10px">
        {{ message }}
      </v-card-text>

      <v-card-actions>
        <v-btn @click="reloadPage()" class="btn">Refrescar página</v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script lang="ts">
import Vue from "vue";
import { FontAwesomeIcon } from "@fortawesome/vue-fontawesome";
import { library } from "@fortawesome/fontawesome-svg-core";
import { faExclamationTriangle } from "@fortawesome/free-solid-svg-icons";
import { Colors } from "../assets/colors";

library.add(faExclamationTriangle);

Vue.component("font-awesome-icon", FontAwesomeIcon);

export default Vue.extend({
  name: "ErrorDialog",
  props: {
    open: Object as () => boolean,
    message: Object as () => string,
  },
  data() {
    return {
      Colors,
    };
  },

  methods: {
    reloadPage() {
      window.location.reload();
    },
  },
});
</script>

<style scoped>
@import url("https://fonts.googleapis.com/css2?family=Poppins&display=swap");

.btn {
  text-transform: capitalize !important;
  color: white !important;
  background: #004e88 !important;
}

.card {
  border-radius: 10px !important;
  font-family: "Poppins", sans-serif;
}

.card-title {
  color: #e41919;
}

.error-dialog {
  box-shadow: none;
}
</style>