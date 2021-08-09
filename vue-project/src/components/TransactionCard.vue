<template>
  <v-sheet elevation="2" class="card">
    <p :style="{ color: Colors.BLUE_TEXT }" class="card-title">
      <b>Transac. No.:</b> {{ transaction.TransactionId }}
    </p>

    <v-divider></v-divider>

    <!-- Transaction details -->
    <v-container class="details-container">
      <v-row align="center" no-gutters>
        <!-- Left col -->
        <v-col class="details-col" style="margin-bottom: auto">
          <v-row align="start" no-gutters>
            <font-awesome-icon
              size="sm"
              style="margin-right: 10px"
              :color="Colors.GREEN"
              :icon="['fas', 'desktop']"
            />

            <p class="device">{{ transaction.Device }}</p>
          </v-row>

          <v-row align="center" no-gutters>
            <font-awesome-icon
              size="sm"
              style="margin: 0px 10px 0px 2px"
              :color="Colors.GREEN"
              :icon="['fas', 'map-marker-alt']"
            />

            <p class="device">{{ transaction.Ip }}</p>
          </v-row>
        </v-col>

        <!-- Right col -->
        <v-col class="details-col" align-self="end">
          <v-row
            style="margin-top: 18px"
            align="bottom"
            no-gutters
            justify="end"
          >
            <font-awesome-icon
              size="lg"
              style="margin-right: 10px"
              :color="Colors.GREEN"
              :icon="['fas', 'dollar-sign']"
            />
            <p style="margin-bottom: 0px; font-size: 18px">
              <b>Total: </b>{{ formatter(getTransactionTotalCost()) }}
            </p>
          </v-row>

          <v-row align="bottom" no-gutters justify="end">
            <v-btn @click="showProducts = !showProducts" class="product-btn">
              <font-awesome-icon
                size="lg"
                style="margin-right: 10px"
                color="white"
                v-bind:icon="
                  showProducts ? ['fas', 'chevron-up'] : ['fas', 'chevron-down']
                "
              />

              Ver Productos
            </v-btn>
          </v-row>
        </v-col>
      </v-row>

      <v-divider v-if="showProducts"></v-divider>

      <div v-if="showProducts" style="margin-top: 10px">
        <v-row
          v-for="product in transaction.Products"
          :key="product.ProductId"
          align="center"
          no-gutters
        >
          <v-col>
            <p class="product-detail">{{ product.Name }}</p>
          </v-col>

          <v-col>
            <v-row no-gutters justify="end">
              <p class="product-detail">{{ formatter(product.Price) }}</p>
            </v-row>
          </v-col>
        </v-row>
      </div>
    </v-container>
  </v-sheet>
</template>

<script lang="ts">
import Vue from "vue";
import { Transaction } from "../types";
import { Colors } from "../assets/colors";
import { FontAwesomeIcon } from "@fortawesome/vue-fontawesome";
import {
  faDesktop,
  faMapMarkerAlt,
  faChevronUp,
  faChevronDown,
  faDollarSign,
} from "@fortawesome/free-solid-svg-icons";
import { library } from "@fortawesome/fontawesome-svg-core";
import { formatter } from "../functions/functions";

library.add(
  faDesktop,
  faChevronUp,
  faChevronDown,
  faMapMarkerAlt,
  faDollarSign
);
Vue.component("font-awesome-icon", FontAwesomeIcon);

export default Vue.extend({
  name: "TrasactionCard",
  data() {
    return {
      Colors,
      formatter,
      showProducts: false,
    };
  },

  props: {
    transaction: Object as () => Transaction,
  },

  methods: {
    getTransactionTotalCost() {
      return this.transaction.Products.map((p) => p.Price).reduce(
        (prev, cur) => prev + cur
      );
    },
  },
});
</script>

<style scoped>
.card {
  border-radius: 10px;
  width: 50%;
  margin: 10px;
  padding: 15px;
}

.card-title {
  margin: 10px 0px 4px 0px !important;
  font-size: 18px !important;
}

.details-container {
}

.details-col {
  width: 50%;
}

.device {
  color: #6c6c6c !important;
  font-size: 14px !important;
  margin-bottom: 0px !important;
  text-transform: capitalize !important;
}

.product-btn {
  background-color: #004e88 !important;
  border-radius: 4px !important;
  color: white !important;
  text-transform: capitalize !important;
  margin-top: 20px;
  margin-bottom: 10px;
}

.product-detail {
  font-size: 15px !important;
  margin-bottom: 0px !important;
}
</style>