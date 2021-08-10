import { Transaction } from "@/types";

export const formatter = Intl.NumberFormat("en-US", {
  style: "currency",
  currency: "USD",
  currencyDisplay: "symbol",
  maximumFractionDigits: 2,
}).format;

export const transaction: Transaction = {
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
};
