<template>
  <div class="main-container">
    <!-- Navbar -->
    <div class="navbar">
      <v-btn text href="/" class="navbar-btn">Inicio</v-btn>
      <v-btn text ref="buyerRef" href="#buyers" class="navbar-btn"
        >Compradores</v-btn
      >
    </div>

    <v-container class="top-container">
      <v-row style="margin-top: 100px" align="center" no-gutters>
        <!-- Title and date -->
        <v-col class="title-col">
          <span style="font-size: 45px; font-weight: bold"
            >Sistema de administración de Restaurante</span
          >

          <v-row no-gutters style="margin-top: 30px">
            <span> Seleccione fecha </span>
          </v-row>

          <v-row no-gutters style="margin-top: 5px">
            <!-- Button -->
            <v-tooltip bottom>
              <template v-slot:activator="{ on, attrs }">
                <v-btn
                  rounded
                  elevation="2"
                  @click="showDatePicker = !showDatePicker"
                  class="calendar-btn"
                  v-bind="attrs"
                  v-on="on"
                >
                  <font-awesome-icon
                    size="lg"
                    style="margin-right: 10px"
                    :color="Colors.BLUE"
                    :icon="['far', 'calendar-alt']"
                  />

                  {{ date }}
                </v-btn>
              </template>
              <span>Seleccione una fecha</span>
            </v-tooltip>

            <v-date-picker
              class="datepicker"
              v-if="showDatePicker"
              v-model="date"
              @change="onDateChange()"
            ></v-date-picker>

            <v-btn
              elevation="2"
              :style="{ 'background-color': Colors.GREEN }"
              @click="loadData()"
              :loading="loadingBuyers"
              class="sync-btn"
            >
              Sincronizar
            </v-btn>
          </v-row>

          <v-row no-gutters style="margin-top: 30px">
            <v-btn
              elevation="2"
              @click="goToBuyers"
              :disabled="loadingBuyers"
              :style="{ color: Colors.BLUE_TEXT }"
              class="buyers-btn"
            >
              Ver compradores
            </v-btn>
          </v-row>
        </v-col>

        <!-- Picture -->
        <v-col class="title-col">
          <v-img
            class="restaurant-img"
            max-height="700"
            max-width="700"
            :src="require('../assets/restaurant.jpg')"
          ></v-img>
        </v-col>
      </v-row>
    </v-container>

    <div class="page-container">
      <h1 id="buyers" :style="{ color: Colors.BLUE }">Compradores</h1>

      <v-data-table
        @click:row="onTableRowClicked"
        :headers="headers"
        :items="buyers"
        :items-per-page="5"
        class="buyers-table"
      ></v-data-table>
    </div>
  </div>
</template>

<script lang="ts">
import Vue from "vue";
import { Colors } from "../assets/colors";
import { FontAwesomeIcon } from "@fortawesome/vue-fontawesome";
import { faCalendarAlt } from "@fortawesome/free-regular-svg-icons";
import { library } from "@fortawesome/fontawesome-svg-core";
import { format } from "date-fns";
import { Buyer } from "../types";

library.add(faCalendarAlt);

Vue.component("font-awesome-icon", FontAwesomeIcon);

export default Vue.extend({
  name: "Home",
  data() {
    return {
      Colors: Colors,
      showDatePicker: false,
      loadingBuyers: false,
      format,
      date: "2020-08-21",
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
      buyers: [] as Buyer[],
    };
  },

  created() {
    this.date = format(Date.now(), "yyyy-MM-DD");
  },

  methods: {
    onDateChange() {
      setTimeout(() => {
        this.showDatePicker = false;
      }, 250);
    },

    onTableRowClicked(item: any, metadata: any) {
      this.$router.push({ path: `/buyer/${item.BuyerId}` });
    },

    goToBuyers() {
      this.fetchBuyers();
      if (this.$refs && this.$refs.buyerRef) {
        this.$refs.buyerRef.$el.click();
      }
    },

    async loadData() {
      this.loadingBuyers = true;

      const res = await fetch("http://localhost:9000/restaurant-data", {
        method: "POST",
        body: JSON.stringify({
          date: this.date,
        }),
      });

      console.log(res);
      this.fetchBuyers();
    },

    async fetchBuyers() {
      this.loadingBuyers = true;
      const res = await fetch("http://localhost:9000/buyer/all", {
        credentials: "include",
      });

      res.json().then((r) => {
        this.loadingBuyers = false;
        let addedBuyers: string[] = [];
        let buyersBuffer: Buyer[] = [];

        r.buyers.forEach((b: any) => {
          if (!addedBuyers.includes(b.BuyerId)) {
            addedBuyers.push(b.BuyerId);
            buyersBuffer.push(b);
          }
        });

        this.buyers = buyersBuffer;
      });
    },
  },
});
</script>

<style >
.buyers-btn {
  background-color: white;
  text-transform: capitalize !important;
  border-radius: 4px !important;
  font-weight: bold !important;
}

.buyers-table {
  margin: 20px 0px 50px 0px;
  width: 50%;
}

.datepicker {
  position: absolute !important;
  top: 370px !important;
  left: 106px !important;
  z-index: 2 !important;
}

.dateText {
  margin-left: 10px;
  border-radius: 50px !important;
  display: flex;
  align-items: center;
  justify-content: center;
}

.calendar-btn {
  background-color: white;
}

.main-container {
  font-family: "Poppins", sans-serif;
}

.navbar {
  display: flex;
  position: absolute;
  margin-top: 12px;
}

.navbar-btn {
  color: white !important;
  margin: 0px 5px !important;
  text-transform: capitalize !important;
}

.page-container {
  width: 90%;
  margin: auto;
}

.restaurant-img {
  border-radius: 10px;
}

.sync-btn {
  margin-left: 20px;
  text-transform: capitalize !important;
  color: white !important;
  font-weight: bold !important;
  border-radius: 4px !important;
}

.table-header {
  font-size: 18px !important;
  color: #004e88 !important;
}

.title-col {
  width: 50%;
}

.top-container {
  margin-bottom: 20px;
  width: 100%;
  padding: 35px !important;
  color: white;
  max-width: 100% !important;
  background: linear-gradient(
    307deg,
    #ffffff,
    #ffffff 50%,
    #004e88 50%,
    #004e88
  );
  height: 100vh !important;
}
</style>

