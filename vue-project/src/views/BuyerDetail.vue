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
      <p v-if="!dataAvailable" class="no-data-text">
        El comprador solicitado no existe.
      </p>

      <div v-if="dataAvailable">
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

          <v-select
            v-model="pageSizeT"
            :items="pageSizeOpts"
            label="Ver"
          ></v-select>

          <v-simple-table>
            <template v-slot:default>
              <thead>
                <tr>
                  <th class="table-header">Número</th>
                  <th class="table-header">Fecha</th>
                  <th class="table-header">Dispositivo</th>
                  <th class="table-header">Ip</th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="transaction in transactions.Transactions"
                  @click="onTransactionClicked(transaction)"
                  :key="transaction.TransactionId"
                >
                  <td>{{ transaction.TransactionId }}</td>
                  <td>{{ transaction["Date"] }}</td>
                  <td>{{ transaction.Device }}</td>
                  <td>{{ transaction.Ip }}</td>
                </tr>
              </tbody>
            </template>
          </v-simple-table>

          <!-- Pagination -->
          <v-pagination
            style="margin-top: 10px"
            v-model="pageT"
            :length="pagLengthT"
          ></v-pagination>

          <v-dialog
            width="700"
            content-class="transaction-dialog"
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

          <v-select
            v-model="pageSizeB"
            :items="pageSizeOpts"
            label="Ver"
          ></v-select>

          <!-- Buyers table -->
          <BuyersTable
            :buyers="buyersWithEqIp.Buyers"
            @pageChange="onBuyersPageChange"
            :page="pageB"
            :pagLength="pagLengthB"
            :pageSize="pageSizeB"
          />
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
            <div
              v-for="product in recommendedProducts"
              :key="product.ProductId"
            >
              <ProductCard :product="product" />
            </div>
          </div>
        </div>
      </div>
    </div>

    <ErrorDialog :open="openErrorDialog" :error="error" />
  </div>
</template>

<script lang="ts">
import Vue from "vue";
import { Colors } from "../assets/colors";
import TrasactionCard from "../components/TransactionCard.vue";
import ProductCard from "../components/ProductCard.vue";
import BuyersTable from "../components/BuyersTable.vue";
import { transaction, dateFormat } from "../functions/functions";
import { Buyer, Product, Transaction } from "../types";
import Axios, { AxiosError } from "axios";
import { handleRequestError } from "../functions/functions";
import { Endpoints } from "../constants/constants";
import ErrorDialog from "../components/ErrorDialog.vue";

export default Vue.extend({
  name: "BuyerDetail",
  components: { TrasactionCard, ProductCard, ErrorDialog, BuyersTable },
  data() {
    return {
      pageSizeOpts: [5, 10, 15, 20],
      pageB: 1,
      pageSizeB: 10,
      pagLengthB: 10,
      pageT: 1,
      pageSizeT: 10,
      pagLengthT: 10,
      Endpoints,
      openTransactionDialog: false,
      loadingBuyerData: true,
      transaction,
      dataAvailable: true,
      error: {
        message: "",
        status: "",
      },
      openErrorDialog: false,
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
      transactions: {
        Transactions: [] as Transaction[],
        Count: Number,
      },
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
      buyersWithEqIp: {
        Buyers: [] as Buyer[],
        Count: Number,
      },
      recommendedProducts: [] as Product[],
    };
  },

  watch: {
    $route() {
      this.fetchBuyer();
    },

    buyersWithEqIp: function () {
      this.pagLengthB = this.getPaginationLengthB();
    },

    transactions: function () {
      this.pagLengthT = this.getPaginationLengthT();
    },

    pageT: function (newVal) {
      this.onTransactionPageChange(newVal);
    },

    pageSizeT: function () {
      this.onTransactionPageChange(1);
    },

    pageSizeB: function () {
      this.onBuyersPageChange(1);
    },
  },

  mounted() {
    this.fetchBuyer();
  },

  methods: {
    onTransactionPageChange(newPage: number) {
      this.pageT = newPage;
      this.fetchBuyer();
    },

    onBuyersPageChange(newPage: number) {
      this.pageB = newPage;
      this.fetchBuyer();
    },

    handleRequestError,

    fetchBuyer() {
      window.scrollTo(0, 0);

      this.loadingBuyerData = true;

      Axios.get(
        `${this.Endpoints.BUYER}/${this.$route.params.id}?pageB=${this.pageB}&pageSizeB=${this.pageSizeB}&pageT=${this.pageT}&pageSizeT=${this.pageSizeT}`,
        {
          withCredentials: true,
        }
      )
        .then((res) => {
          this.transactions = this.parseTransactions(
            res.data.TransactionHistory
          );

          if (this.transactions.Transactions.length === 0) {
            this.dataAvailable = false;
          }

          this.buyersWithEqIp = res.data.BuyersWithSameIp;
          this.recommendedProducts = res.data.RecommendedProducts;
          this.loadingBuyerData = false;
        })
        .catch((error: AxiosError) => {
          this.loadingBuyerData = false;
          this.error = this.handleRequestError(error);
          this.openErrorDialog = true;
        });
    },

    /**
     * Converts the date to 'DD/MM/yyyy' format and
     * capitalizes the 'device' field.
     */
    parseTransactions(transactionHistory: any) {
      let array = transactionHistory.Transactions.map((t: any) => {
        let Device = t.Device.slice(0, 1).toUpperCase() + t.Device.slice(1);
        let date = dateFormat(new Date(t.Date));

        return { ...t, Device, Date: date };
      });

      return { ...transactionHistory, Transactions: array };
    },

    /**
     * Filters out repeated buyers and the currently seen buyer.
     */
    filterBuyers(buyersWithSameIp: any): any {
      let addedBuyers: string[] = [];
      let buyersBuffer: Buyer[] = [];

      buyersWithSameIp.Buyers.forEach((b: any) => {
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

      buyersWithSameIp = { ...buyersWithSameIp, Buyers: buyersBuffer };

      return buyersWithSameIp;
    },

    closeDialog() {
      this.openTransactionDialog = false;
    },

    onTransactionClicked(item: any) {
      this.openTransactionDialog = true;
      this.transaction = this.transactions.Transactions.filter(
        (t) => t.TransactionId === item.TransactionId
      )[0];
    },

    onBuyerClicked(item: any) {
      this.$router.push({ path: `/buyer/${item.BuyerId}` });
    },

    getPaginationLengthB() {
      return Math.ceil(this.buyersWithEqIp.Count / this.pageSizeB);
    },

    getPaginationLengthT() {
      return Math.ceil(this.transactions.Count / this.pageSizeT);
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

.no-data-text {
  width: 50%;
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

table tr {
  cursor: pointer;
}

.transaction-dialog {
  box-shadow: none;
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