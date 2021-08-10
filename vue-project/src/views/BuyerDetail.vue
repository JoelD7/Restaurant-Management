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
      <h2 :style="{ color: Colors.BLUE_TEXT }">
        Comprador: <span style="font-weight: normal">{{ getBuyerName() }}</span>
      </h2>

      <!-- Transaction History -->
      <div>
        <h1 :style="{ color: Colors.BLUE }">Historial de Transacciones</h1>

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
        <h1 id="buyers" :style="{ color: Colors.BLUE }">Compradores</h1>

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
          style="width: 70%"
        ></v-data-table>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { parseISO } from "date-fns";
import Vue from "vue";
import { Colors } from "../assets/colors";
import TrasactionCard from "../components/TransactionCard.vue";
import { transaction, dateFormat } from "../functions/functions";

export default Vue.extend({
  name: "BuyerDetail",
  components: { TrasactionCard },
  data() {
    return {
      openTransactionDialog: false,
      loadingBuyerData: true,
      transaction,
      dateFormat,
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
      transactions: [],
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
      buyersWithEqIp: [],
    };
  },

  async mounted() {
    const res = await fetch(
      `http://localhost:9000/buyer/${this.$route.params.id}`
    );

    res.json().then((r) => {
      this.transactions = r.TransactionHistory.map((t: any) => {
        let Device = t.Device.slice(0, 1).toUpperCase() + t.Device.slice(1);
        let date = dateFormat(new Date(t.Date));

        return { ...t, Device, Date: date };
      });

      this.buyersWithEqIp = r.BuyersWithSameIp;
      this.loadingBuyerData = false;
    });
  },

  methods: {
    getBuyerName() {
      if (localStorage.getItem("buyerName") !== null) {
        return localStorage.getItem("buyerName");
      }

      return "";
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
      localStorage.setItem("buyerName", item.Name);
      window.location.reload();
    },
    parseISO,
  },
});
</script>

<style >
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
</style>