<template>
  <div class="main-container">
    <!-- Navbar -->
    <div class="navbar">
      <v-btn text href="/" class="navbar-btn">Inicio</v-btn>
      <v-btn text @click="seeBuyers" class="navbar-btn">Compradores</v-btn>
    </div>

    <v-container class="top-container">
      <v-row style="margin-top: 100px" align="center" no-gutters>
        <!-- Title and date -->
        <v-col
          class="title-col"
          :cols="$vuetify.breakpoint.width < 1085 ? 12 : 6"
        >
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
                  @click="showDatePicker = true"
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

            <v-dialog
              width="500"
              content-class="datepicker-dialog"
              @click:outside="showDatePicker = false"
              v-model="showDatePicker"
            >
              <v-date-picker
                class="datepicker"
                v-if="showDatePicker"
                v-model="date"
                :max="maxDate"
                @change="onDateChange()"
              ></v-date-picker>
            </v-dialog>

            <v-btn
              elevation="2"
              :style="{ 'background-color': Colors.GREEN }"
              @click="loadData()"
              :loading="loadingBuyers"
              :disabled="loadingBuyers"
              class="sync-btn"
            >
              Sincronizar
            </v-btn>
          </v-row>

          <v-row no-gutters style="margin-top: 30px">
            <v-btn
              elevation="2"
              @click="seeBuyers"
              :disabled="loadingBuyers"
              :style="{ color: Colors.BLUE_TEXT }"
              class="buyers-btn"
            >
              Ver compradores
            </v-btn>
          </v-row>
        </v-col>

        <!-- Picture -->
        <v-col
          class="title-col"
          :cols="$vuetify.breakpoint.width < 1085 ? 0 : 6"
        >
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
      <h1 :style="{ color: Colors.BLUE }">Compradores</h1>

      <p v-if="!dataAvailable" class="no-data-text">
        La base de datos no cuenta con información para satisfacer esta
        solicitud. Seleccione una fecha y luego presione el botón de
        <b :style="{ color: Colors.GREEN }">Sincronizar</b> para mostrar la
        lista de compradores.
      </p>

      <v-data-table
        @click:row="onTableRowClicked"
        v-if="dataAvailable"
        :headers="headers"
        :items="buyers"
        :items-per-page="10"
        class="buyers-table"
      ></v-data-table>

      <ErrorDialog :open="openErrorDialog" :error="error" />
    </div>

    <v-snackbar
      :timeout="4000"
      :value="openSnackbar"
      absolute
      top
      :color="Colors.ORANGE"
      elevation="24"
    >
      Esta fecha ya ha sido sincronizada.
    </v-snackbar>
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
import { Endpoints } from "../constants/constants";
import Axios, { AxiosError } from "axios";
import ErrorDialog from "../components/ErrorDialog.vue";
import { handleRequestError } from "../functions/functions";

library.add(faCalendarAlt);

Vue.component("font-awesome-icon", FontAwesomeIcon);

export default Vue.extend({
  name: "Home",
  data() {
    return {
      Endpoints,
      error: {
        message: "",
        status: "",
      },
      openSnackbar: false,
      Colors,
      showDatePicker: false,
      dataAvailable: true,
      maxDate: "",
      openErrorDialog: false,
      openTransactionDialog: false,
      loadingBuyers: false,
      format,
      date: "",
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

  components: { ErrorDialog },

  mounted() {
    this.fetchBuyers();
  },

  created() {
    let d = new Date(Date.now());
    let month: number = d.getMonth() + 1;

    this.date = `${d.getFullYear()}-${
      month < 10 ? "0" + month : month
    }-${d.getDate()}`;

    this.maxDate = `${d.getFullYear()}-${
      month < 10 ? "0" + month : month
    }-${d.getDate()}`;
  },

  methods: {
    onDateChange() {
      setTimeout(() => {
        this.showDatePicker = false;
      }, 250);
    },

    onTableRowClicked(item: any) {
      this.$router.push({ path: `/buyer/${item.BuyerId}` });
    },

    seeBuyers() {
      if (this.buyers.length === 0) {
        this.fetchBuyers();
      }
      window.scrollTo(0, document.body.scrollHeight);
    },

    loadData() {
      this.loadingBuyers = true;
      this.openSnackbar = false;

      Axios.post(
        this.Endpoints.RESTAURANT_DATA,
        {
          date: this.date,
        },
        { withCredentials: true }
      )
        .then((r) => {
          /**
           * If no data is returned then it means the date has already been synched.
           */
          if (r.data !== "") {
            this.fetchBuyers();
          } else {
            this.openSnackbar = true;
            this.loadingBuyers = false;
          }
        })
        .catch((error: AxiosError) => {
          this.error = this.handleRequestError(error);
          this.openErrorDialog = true;
          this.loadingBuyers = false;
        });
    },

    fetchBuyers() {
      this.loadingBuyers = true;

      Axios.get(this.Endpoints.ALL_BUYERS, { withCredentials: true })
        .then((res) => {
          this.loadingBuyers = false;
          this.buyers = res.data.buyers;

          if (this.buyers.length === 0) {
            this.dataAvailable = false;
          } else {
            this.dataAvailable = true;
          }
        })
        .catch((error: AxiosError) => {
          this.error = this.handleRequestError(error);
          this.openErrorDialog = true;
        });
    },

    handleRequestError,
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

.buyers-btn:disabled {
  background-color: #ffffffcc !important;
}

.buyers-table {
  margin: 20px 0px 50px 0px;
  width: 50%;
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

.datepicker {
  margin: auto;
  width: fit-content !important;
}

.datepicker-dialog {
  box-shadow: none !important;
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

.no-data-text {
  width: 50%;
}

@media (max-width: 1250px) {
  .no-data-text {
    width: 65%;
  }
}

@media (max-width: 780px) {
  .no-data-text {
    width: 85%;
  }
}

@media (max-width: 650px) {
  .no-data-text {
    width: 100%;
  }
}

.page-container {
  width: 90%;
  margin: auto;
}

.restaurant-img {
  border-radius: 10px;
}

@media (max-width: 1270px) {
  .restaurant-img {
    border-radius: 10px;
    max-height: 600px !important;
    max-width: 600px !important;
  }
}

@media (max-width: 1140px) {
  .restaurant-img {
    max-height: 500px !important;
    max-width: 500px !important;
  }
}

@media (max-width: 1085px) {
  .restaurant-img {
    display: none !important;
  }
}

.sync-btn {
  margin-left: 20px;
  text-transform: capitalize !important;
  color: white !important;
  font-weight: bold !important;
  border-radius: 4px !important;
}

.sync-btn:disabled {
  color: #ffffffbd !important;
  background-color: #78bb1ea8 !important;
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
  min-height: 100vh !important;
}

@media (max-width: 1085px) {
  .top-container {
    min-height: 55vh !important;
    background: #004e88;
  }
}
</style>

