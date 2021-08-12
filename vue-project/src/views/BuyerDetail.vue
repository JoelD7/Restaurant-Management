<template>
  <div class="main-container">
    <!-- Navbar -->
    <div class="navbar-buyer">
      <v-btn text href="/" class="navbar-btn-buyer">Inicio</v-btn>
      <v-btn text ref="buyerRef" href="#buyers" class="navbar-btn-buyer"
        >Compradores</v-btn
      >
    </div>

    <div class="page-container">
      <h3 :style="{ color: Colors.BLUE_TEXT }">
        Comprador: <span style="font-weight: normal">{{ buyerName }}</span>
      </h3>

      <!-- Transaction History -->
      <div>
        <h2 :style="{ color: Colors.BLUE }">Historial de Transacciones</h2>

        <div class="progress-container">
          <v-progress-circular
            :size="70"
            v-if="loadingBuyerData"
            indeterminate
            :color="Colors.GREEN"
          ></v-progress-circular>
        </div>

        <v-data-table
          v-if="!loadingBuyerData"
          @click:row="onTransactionClicked"
          :headers="headers"
          :items="transactions"
          :items-per-page="10"
          class="transactions-table"
        ></v-data-table>

        <v-dialog
          width="700"
          @click:outside="closeDialog"
          v-model="openTransactionDialog"
        >
          <TrasactionCard
            v-if="openTransactionDialog"
            :transaction="transaction"
          />
        </v-dialog>
      </div>

      <!-- Buyers with equal IP -->
      <div style="margin-top: 40px">
        <h2 id="buyers" :style="{ color: Colors.BLUE }">Compradores</h2>

        <div class="progress-container">
          <v-progress-circular
            :size="70"
            v-if="loadingBuyerData"
            indeterminate
            :color="Colors.GREEN"
          ></v-progress-circular>
        </div>

        <v-data-table
          v-if="!loadingBuyerData"
          @click:row="onBuyerClicked"
          :headers="buyerHeaders"
          :items="buyersWithEqIp"
          :items-per-page="5"
          class="buyers-table"
        ></v-data-table>
      </div>

      <!-- Recommended products -->
      <div style="margin-top: 40px">
        <h2 :style="{ color: Colors.BLUE }">
          Productos que podrian interesar a {{ buyerName }}
        </h2>

        <div class="progress-container">
          <v-progress-circular
            :size="70"
            v-if="loadingBuyerData"
            indeterminate
            :color="Colors.GREEN"
          ></v-progress-circular>
        </div>

        <div v-if="!loadingBuyerData" class="product-card-container">
          <div v-for="product in recommendedProducts" :key="product.ProductId">
            <ProductCard :product="product" />
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import Vue from "vue";
import { Colors } from "../assets/colors";
import TrasactionCard from "../components/TransactionCard.vue";
import ProductCard from "../components/ProductCard.vue";
import { transaction, dateFormat } from "../functions/functions";
import { Buyer, Product, Transaction } from "../types";
import { Endpoints } from "../constants/constants";

export default Vue.extend({
  name: "BuyerDetail",
  components: { TrasactionCard, ProductCard },
  data() {
    return {
      Endpoints,
      openTransactionDialog: false,
      loadingBuyerData: true,
      transaction,
      dateFormat,
      buyerName: "",
      Colors,
      transactionsBuffer: [],
      headers: [
        {
          text: "Número",
          value: "TransactionId",
          class: "transaction-table-header",
        },
        {
          text: "Fecha",
          value: "Date",
          class: "transaction-table-header",
        },
        {
          text: "Dispositivo",
          value: "Device",
          class: "transaction-table-header",
        },
        {
          text: "Ip",
          value: "Ip",
          class: "transaction-table-header",
        },
      ],
      transactions: [] as Transaction[],
      buyerHeaders: [
        {
          text: "ID",
          value: "BuyerId",
          class: "transaction-table-header",
        },
        {
          text: "Nombre",
          value: "Name",
          class: "transaction-table-header",
        },
        {
          text: "Edad",
          value: "Age",
          class: "transaction-table-header",
        },
      ],
      buyersWithEqIp: [] as Buyer[],
      recommendedProducts: [] as Product[],
    };
  },

  watch: {
    $route() {
      this.fetchBuyer();
    },
  },

  async mounted() {
    this.fetchBuyer();
  },

  methods: {
    async fetchBuyer() {
      window.scrollTo(0, 0);

      this.loadingBuyerData = true;

      const res = await fetch(
        `${this.Endpoints.BUYER}/${this.$route.params.id}`
      );

      res.json().then((r) => {
        this.transactions = this.parseTransactions(r.TransactionHistory);
        this.buyersWithEqIp = this.filterBuyers(r.BuyersWithSameIp);
        this.recommendedProducts = r.RecommendedProducts;
        this.loadingBuyerData = false;
      });
    },

    /**
     * Converts the date to 'DD/MM/yyyy' format and
     * capitalizes the 'device' field.
     */
    parseTransactions(transactionHistory: any) {
      return transactionHistory.map((t: any) => {
        let Device = t.Device.slice(0, 1).toUpperCase() + t.Device.slice(1);
        let date = dateFormat(new Date(t.Date));

        return { ...t, Device, Date: date };
      });
    },

    /**
     * Filters out repeated buyers and the currently seen buyer.
     */
    filterBuyers(buyersWithSameIp: any): Buyer[] {
      let addedBuyers: string[] = [];
      let buyersBuffer: Buyer[] = [];

      buyersWithSameIp.forEach((b: any) => {
        if (
          !addedBuyers.includes(b.BuyerId) &&
          b.BuyerId !== this.$route.params.id
        ) {
          addedBuyers.push(b.BuyerId);
          buyersBuffer.push(b);
        }

        if (b.BuyerId === this.$route.params.id) {
          this.buyerName = b.Name;
        }
      });

      return buyersBuffer;
    },

    closeDialog() {
      this.openTransactionDialog = false;
    },

    onTransactionClicked(item: any) {
      this.openTransactionDialog = true;
      this.transaction = this.transactions.filter(
        (t) => t.TransactionId === item.TransactionId
      )[0];
    },

    onBuyerClicked(item: any) {
      this.$router.push({ path: `/buyer/${item.BuyerId}` });
    },
  },
});
</script>

<style >
.buyers-table {
  margin: 20px 0px 50px 0px;
  width: 70%;
}

@media (max-width: 960px) {
  .buyers-table {
    width: 75%;
  }
}

@media (max-width: 800px) {
  .buyers-table {
    width: 85%;
  }
}

@media (max-width: 640px) {
  .buyers-table {
    width: 100%;
  }
}

.main-container {
  font-family: "Poppins", sans-serif;
}

.navbar-buyer {
  display: flex;
  margin-top: 12px;
}

.navbar-btn-buyer {
  color: #004e88 !important;
  margin: 0px 5px !important;
  text-transform: capitalize !important;
}

.page-container {
  width: 90%;
  margin: 20px auto 50px auto;
}

.product-card-container {
  display: flex;
  flex-flow: wrap;
}

.progress-container {
  display: flex;
  justify-content: center;
  margin-top: 20px;
  width: 30%;
}

.transaction-table-header {
  font-size: 18px !important;
  color: #004e88 !important;
}

.transactions-table {
  width: 70%;
}

@media (max-width: 960px) {
  .transactions-table {
    width: 80%;
  }
}

@media (max-width: 800px) {
  .transactions-table {
    width: 90%;
  }
}

@media (max-width: 640px) {
  .transactions-table {
    width: 100%;
  }
}
</style>