<template>
  <div class="main-container">
    <v-container class="top-container">
      <v-row style="margin-top: 100px" align="center" no-gutters>
        <v-col class="title-col">
          <span style="font-size: 45px; font-weight: bold"
            >Sistema de administración de Restaurante</span
          >

          <v-row no-gutters style="margin-top: 30px">
            <span> Seleccione fecha </span>
          </v-row>

          <v-row no-gutters>
            <v-tooltip bottom>
              <template v-slot:activator="{ on, attrs }">
                <v-btn rounded class="calendar-btn" v-bind="attrs" v-on="on">
                  <font-awesome-icon
                    size="lg"
                    :color="Colors.BLUE"
                    :icon="['far', 'calendar-alt']"
                  />
                </v-btn>
              </template>
              <span>Select date</span>
            </v-tooltip>
          </v-row>
        </v-col>

        <v-col class="title-col">b</v-col>
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
    };
  },
});
</script>

<style>
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

