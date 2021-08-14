import { Transaction, CustomError } from "@/types";
import { AxiosError } from "axios";

export const currencyFormatter = Intl.NumberFormat("en-US", {
  style: "currency",
  currency: "USD",
  currencyDisplay: "symbol",
  maximumFractionDigits: 2,
}).format;

export const dateFormat = Intl.DateTimeFormat("es-ES", {
  year: "numeric",
  month: "2-digit",
  day: "2-digit",
  timeZone: "UTC",
}).format;

export const transaction: Transaction = {
  TransactionId: "00005f39cef1",
  BuyerId: "2d8e2eb5",
  Ip: "211.133.165.230",
  Device: "linux",
  Date: "2020-08-17T00:00:00Z",
  Products: ["cd3de2cc", "4bb66fdd"],
};

export function handleRequestError(error: AxiosError): CustomError {
  if (error.response) {
    return {
      message: error.response.data,
      status: String(error.response.status),
    };
  } else if (error.request) {
    return {
      message:
        "La solicitud realizada al servidor no fue contestada. Al parecer " +
        "el servidor no se encuentra disponible. Por favor intente más tarde. ",
      status: error.code ? error.code : "",
    };
  } else {
    return {
      message: error.message,
      status: error.code ? error.code : "",
    };
  }
}
