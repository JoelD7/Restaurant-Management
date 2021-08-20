<template>
  <div class="main-container">
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
            v-for="transaction in transactions"
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
      :total-visible="5"
      v-model="page"
      :length="pagLength"
    ></v-pagination>

    <v-dialog
      width="700"
      content-class="transaction-dialog"
      @click:outside="closeDialog"
      v-model="openTransactionDialog"
    >
      <TrasactionCard v-if="openTransactionDialog" :transaction="transaction" />
    </v-dialog>
  </div>
</template>

<script lang="ts">
import { Transaction } from "@/types";
import Vue from "vue";
import TrasactionCard from "./TransactionCard.vue";
import { transaction } from "../functions/functions";

export default Vue.extend({
  name: "TransactionsTable",
  props: {
    transactions: Array,
    page: Number,
    pagLength: Number,
    pageSize: Number,
  },
  components: { TrasactionCard },

  data() {
    return {
      transaction,
      openTransactionDialog: false,
    };
  },

  watch: {
    page: function (newVal) {
      this.$emit("pageChange", newVal);
    },
  },

  methods: {
    closeDialog() {
      this.openTransactionDialog = false;
    },

    onTransactionClicked(item: any) {
      this.openTransactionDialog = true;
      let buf = this.transactions.filter(
        (t: any) => t.TransactionId === item.TransactionId
      )[0];

      this.transaction = buf as Transaction;
    },
  },
});
</script>

<style scoped>
</style>