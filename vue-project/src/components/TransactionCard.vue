<template>
  <v-sheet elevation="2" class="card">
    <p :style="{ color: Colors.BLUE_TEXT }" class="card-title">
      <b>Transac. No.:</b> {{ transaction.TransactionId }}
    </p>

    <v-divider></v-divider>

    <!-- Transaction details -->
    <v-container class="details-container">
      <v-row align="center" no-gutters>
        <v-col class="details-col">
          <v-row align="center" no-gutters>
            <font-awesome-icon
              size="lg"
              style="margin-right: 10px"
              :color="Colors.GREEN"
              :icon="['fas', 'fa-desktop']"
            />

            <p class="device">{{ transaction.Device }}</p>
          </v-row>
        </v-col>

        <v-col class="details-col"></v-col>
      </v-row>
    </v-container>
  </v-sheet>
</template>

<script lang="ts">
import Vue from "vue";
import { Transaction } from "../types";
import { Colors } from "../assets/colors";
import { FontAwesomeIcon } from "@fortawesome/vue-fontawesome";
import { faDesktop } from "@fortawesome/free-solid-svg-icons";
Vue.component("font-awesome-icon", FontAwesomeIcon);

export default Vue.extend({
  name: "TrasactionCard",
  data() {
    return {
      Colors,
    };
  },

  props: {
    transaction: Object as () => Transaction,
  },
});
</script>

<style scoped>
.card {
  border-radius: 10px;
  width: 50%;
  margin: 10px;
  padding: 10px;
}

.card-title {
  margin-bottom: 4px !important;
  font-size: 18px !important;
}

.details-container {
}

.details-col {
  width: 50%;
}

.device {
  color: #6c6c6c !important;
  text-transform: capitalize !important;
}
</style>