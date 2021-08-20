<template>
  <div class="main-container">
    <v-simple-table>
      <template v-slot:default>
        <thead>
          <tr>
            <th class="table-header">ID</th>
            <th class="table-header">Nombre</th>
            <th class="table-header">Edad</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="buyer in buyers"
            @click="onBuyerClick(buyer)"
            :key="buyer.BuyerId"
          >
            <td>{{ buyer.BuyerId }}</td>
            <td>{{ buyer.Name }}</td>
            <td>{{ buyer.Age }}</td>
          </tr>
        </tbody>
      </template>
    </v-simple-table>

    <!-- Pagination -->
    <v-pagination
      style="margin-top: 10px"
      v-model="page"
      :total-visible="5"
      :length="pagLength"
    ></v-pagination>
  </div>
</template>

<script lang="ts">
import Vue from "vue";

export default Vue.extend({
  name: "BuyersTable",
  props: {
    buyers: Array,
    page: Number,
    pagLength: Number,
    pageSize: Number,
  },
  watch: {
    page: function (newVal) {
      this.$emit("pageChange", newVal);
    },
  },
  data() {
    return {
      headers: [
        {
          text: "ID",
          value: "BuyerId",
          class: "table-header",
        },
        {
          text: "Nombre",
          value: "Name",
          class: "table-header",
        },
        {
          text: "Edad",
          value: "Age",
          class: "table-header",
        },
      ],
    };
  },

  methods: {
    pageChange(event: any) {
      console.log(event);
    },

    onBuyerClick(item: any) {
      this.$router.push({ path: `/buyer/${item.BuyerId}` });
    },
  },
});
</script>

<style scoped>
.main-container {
  font-family: "Poppins", sans-serif;
}

.table-header {
  font-size: 18px !important;
  color: #004e88 !important;
}

.page-size-select {
  width: 100px !important;
}
</style>