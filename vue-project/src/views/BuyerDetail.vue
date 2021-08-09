<template>
  <div class="main-container">
    <!-- Navbar -->
    <div class="navbar">
      <v-btn text href="/" class="navbar-btn">Inicio</v-btn>
      <v-btn text ref="buyerRef" href="#buyers" class="navbar-btn"
        >Compradores</v-btn
      >
    </div>

    <div class="page-container">
      <h1 :style="{ color: Colors.BLUE }">Historial de Transacciones</h1>

      <div v-for="transaction in transactions" :key="transaction.TransactionId">
        <TrasactionCard :transaction="transaction" />
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import Vue from "vue";
import { Colors } from "../assets/colors";
import TrasactionCard from "../components/TransactionCard.vue";

export default Vue.extend({
  name: "BuyerDetail",
  components: { TrasactionCard },
  data() {
    return {
      Colors,
      transactions: [
        {
          TransactionId: "00005f39cef1",
          BuyerId: "2d8e2eb5",
          Ip: "211.133.165.230",
          Device: "linux",
          Date: "2020-08-17T00:00:00Z",
          Products: [
            {
              ProductId: "cd3de2cc",
              Name: "Fully cooked ready pasta",
              Date: "2020-08-17T00:00:00Z",
              Price: 5449,
            },
            {
              ProductId: "4bb66fdd",
              Name: "Original mashed potatoes",
              Date: "2020-08-17T00:00:00Z",
              Price: 7506,
            },
          ],
        },
        {
          TransactionId: "00005f39d084",
          BuyerId: "279917bb",
          Ip: "126.151.131.211",
          Device: "linux",
          Date: "2020-08-17T00:00:00Z",
          Products: [
            {
              ProductId: "44b04768",
              Name: "Fat Free Refried Beans",
              Date: "2020-08-17T00:00:00Z",
              Price: 3741,
            },
            {
              ProductId: "8324b3dc",
              Name: "Vegan thai coconut big bowl of noodles",
              Date: "2020-08-17T00:00:00Z",
              Price: 986,
            },
            {
              ProductId: "e70d94f9",
              Name: "Vegan noodle",
              Date: "2020-08-17T00:00:00Z",
              Price: 7192,
            },
          ],
        },
        {
          TransactionId: "00005f39d1d0",
          BuyerId: "37a17758",
          Ip: "82.89.238.170",
          Device: "android",
          Date: "2020-08-17T00:00:00Z",
          Products: [
            {
              ProductId: "e835860f",
              Name: "Progresso Traditional Cheese Tortellini in Garden Vegetable Tomato Soup",
              Date: "2020-08-17T00:00:00Z",
              Price: 2061,
            },
            {
              ProductId: "90a1e574",
              Name: "Bûche fromage de Chèvre",
              Date: "2020-08-17T00:00:00Z",
              Price: 9421,
            },
            {
              ProductId: "bae389d5",
              Name: '"Campbell',
              Date: "2020-08-17T00:00:00Z",
              Price: 0,
            },
          ],
        },
      ],
    };
  },
});
</script>

<style scoped>
.main-container {
  font-family: "Poppins", sans-serif;
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

.page-container {
  width: 90%;
  margin: 20px auto 50px auto;
}
</style>