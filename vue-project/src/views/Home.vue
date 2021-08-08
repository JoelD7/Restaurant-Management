<template>
  <div class="main-container">
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
                    :color="Colors.BLUE"
                    :icon="['far', 'calendar-alt']"
                  />
                </v-btn>
              </template>
              <span>Seleccione una fecha</span>
            </v-tooltip>

            <v-date-picker
              class="datepicker"
              v-if="showDatePicker"
              v-model="date"
              @change="onChange()"
            ></v-date-picker>

            <v-sheet
              color="white"
              elevation="2"
              class="dateText"
              height="36"
              width="200"
            >
            </v-sheet>

            <v-btn
              elevation="2"
              :style="{ 'background-color': Colors.GREEN }"
              class="sync-btn"
            >
              Sincronizar
            </v-btn>
          </v-row>

          <v-row no-gutters style="margin-top: 30px">
            <v-btn
              elevation="2"
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
  </div>
</template>

<script lang="ts">
import Vue from "vue";
import { Colors } from "../assets/colors";
import { FontAwesomeIcon } from "@fortawesome/vue-fontawesome";
import { faCalendarAlt } from "@fortawesome/free-regular-svg-icons";
import { library } from "@fortawesome/fontawesome-svg-core";

library.add(faCalendarAlt);

Vue.component("font-awesome-icon", FontAwesomeIcon);

export default Vue.extend({
  name: "Home",
  data() {
    return {
      Colors: Colors,
      showDatePicker: false,
      date: "2021-08-21",
    };
  },

  methods: {
    onChange() {
      console.log(this.date);
    },
  },
});
</script>

<style>
.buyers-btn {
  background-color: white;
  text-transform: capitalize !important;
  border-radius: 4px !important;
  font-weight: bold !important;
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
}

.calendar-btn {
  background-color: white;
}

.main-container {
  font-family: "Poppins", sans-serif;
  background: linear-gradient(
    307deg,
    #ffffff,
    #ffffff 50%,
    #004e88 50%,
    #004e88
  );
  height: 100vh;
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

.title-col {
  width: 50%;
}

.top-container {
  width: 100%;
  padding: 35px !important;
  color: white;
  max-width: 100% !important;
}
</style>

