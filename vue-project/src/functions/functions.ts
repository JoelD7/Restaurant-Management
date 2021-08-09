export const formatter = Intl.NumberFormat("en-US", {
  style: "currency",
  currency: "USD",
  currencyDisplay: "symbol",
  maximumFractionDigits: 2,
}).format;
