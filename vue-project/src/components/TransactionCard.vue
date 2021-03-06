<template>
  <div>
    <v-sheet v-if="!openErrorDialog" elevation="2" class="card">
      <h2 :style="{ color: Colors.BLUE_TEXT }" class="card-title">
        Transac. No.:
        <span style="font-weight: normal">{{ transaction.TransactionId }}</span>
      </h2>

      <v-divider></v-divider>

      <v-container>
        <!-- Transaction details -->
        <v-row align="center" no-gutters>
          <!-- Left col -->
          <v-col class="details-col" style="margin-bottom: auto">
            <!-- Device -->
            <v-row align="start" no-gutters>
              <font-awesome-icon
                size="sm"
                style="margin-right: 10px"
                :color="Colors.GREEN"
                :icon="['fas', 'desktop']"
              />

              <p class="device">{{ transaction.Device }}</p>
            </v-row>

            <!-- IP -->
            <v-row align="center" no-gutters>
              <font-awesome-icon
                size="sm"
                style="margin: 0px 10px 0px 2px"
                :color="Colors.GREEN"
                :icon="['fas', 'map-marker-alt']"
              />

              <p class="device">{{ transaction.Ip }}</p>
            </v-row>

            <!-- Date -->
            <v-row align="center" no-gutters>
              <font-awesome-icon
                size="sm"
                style="margin: 0px 10px 0px 2px"
                :color="Colors.GREEN"
                :icon="['far', 'calendar-alt']"
              />

              <p class="device">
                {{ transaction["Date"] }}
              </p>
            </v-row>
          </v-col>
        </v-row>

        <!-- Productos -->
        <div style="margin-top: 30px">
          <h3 :style="{ color: Colors.BLUE }">Productos</h3>

          <v-progress-circular
            v-if="loadingProducts"
            indeterminate
            :color="Colors.GREEN"
          ></v-progress-circular>

          <div
            v-for="product in products"
            :key="product.ProductId"
            style="margin: 10px 0px"
          >
            <v-row align="center" no-gutters>
              <v-col>
                <p class="product-detail">{{ product.Name }}</p>
              </v-col>

              <v-col>
                <v-row no-gutters justify="end">
                  <p class="product-detail">
                    {{ currencyFormatter(product.Price) }}
                  </p>
                </v-row>
              </v-col>
            </v-row>

            <v-divider></v-divider>
          </div>
        </div>

        <!-- Products Total Cost -->
        <v-row style="margin-top: 18px" align="bottom" no-gutters justify="end">
          <font-awesome-icon
            size="lg"
            style="margin-right: 10px"
            :color="Colors.GREEN"
            :icon="['fas', 'dollar-sign']"
          />
          <p style="margin-bottom: 0px; font-size: 18px">
            <b>Total: </b>{{ currencyFormatter(getTransactionTotalCost()) }}
          </p>
        </v-row>
      </v-container>
    </v-sheet>

    <ErrorDialog :open="openErrorDialog" :error="error" />
  </div>
</template>

<script lang="ts">
import Vue from "vue";
import { Transaction, Product } from "../types";
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
import { faCalendarAlt } from "@fortawesome/free-regular-svg-icons";
import { currencyFormatter } from "../functions/functions";
import { Endpoints } from "../constants/constants";
import Axios, { AxiosError } from "axios";
import { handleRequestError } from "../functions/functions";
import ErrorDialog from "../components/ErrorDialog.vue";

library.add(
  faDesktop,
  faChevronUp,
  faChevronDown,
  faMapMarkerAlt,
  faDollarSign,
  faCalendarAlt
);
Vue.component("font-awesome-icon", FontAwesomeIcon);

export default Vue.extend({
  name: "TrasactionCard",
  components: { ErrorDialog },
  data() {
    return {
      Endpoints,
      Colors,
      currencyFormatter,
      showProducts: false,
      error: {
        message: "",
        status: "",
      },
      openErrorDialog: false,
      loadingProducts: false,
      products: [] as Product[],
    };
  },

  props: {
    transaction: Object as () => Transaction,
  },

  async mounted() {
    this.loadingProducts = true;
    let productIds = this.transaction.Products.join(",");

    Axios.get(`${this.Endpoints.PRODUCTS}?products=${productIds}`, {
      withCredentials: true,
    })
      .then((res) => {
        this.loadingProducts = false;
        this.products = res.data.products;
      })
      .catch((error: AxiosError) => {
        this.error = this.handleRequestError(error);
        this.openErrorDialog = true;
      });
  },

  methods: {
    handleRequestError,
    getTransactionTotalCost() {
      return this.products
        .map((p) => p.Price)
        .reduce((prev, cur) => prev + cur);
    },
  },
});
</script>

<style scoped>
.card {
  border-radius: 10px;
  width: 95%;
  margin: 10px;
  padding: 15px;
}

.card-title {
  margin: 10px 0px 4px 0px !important;
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