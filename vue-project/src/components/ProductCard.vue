<template>
  <v-sheet
    elevation="3"
    class="card"
    :style="{ 'background-color': Colors.GREEN }"
  >
    <h4 style="color: white; font-weight: normal">{{ product.Name }}</h4>

    <v-divider class="divider"></v-divider>

    <h4 class="price">
      {{ currencyFormatter(product.Price) }}
    </h4>
  </v-sheet>
</template>

<script lang="ts">
import Vue from "vue";
import { Colors } from "../assets/colors";
import { currencyFormatter } from "../functions/functions";
import { Product } from "../types";

export default Vue.extend({
  name: "ProductCard",
  data() {
    return {
      Colors,
      currencyFormatter,
    };
  },
  props: {
    product: Object as () => Product,
  },
});
</script>

<style scoped>
.card {
  border-radius: 10px;
  margin: 5px;
  padding: 10px;
  width: 245px;
}

.divider {
  background-color: white;
  margin: 5px 0px;
}

.price {
  color: white;
  text-align: end;
  font-weight: normal;
}
</style>