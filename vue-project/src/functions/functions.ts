import { Transaction } from "@/types";

export const formatter = Intl.NumberFormat("en-US", {
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
