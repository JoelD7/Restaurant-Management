export type Transaction = {
  TransactionId: string;
  BuyerId: string;
  Ip: string;
  Device: string;
  Date: string;
  Products: string[];
};

export interface Product {
  ProductId: string;
  Name: string;
  Date: string;
  Price: number;
}
