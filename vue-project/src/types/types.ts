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

export interface Buyer {
  BuyerId: string;
  Name: string;
  Date: string;
  Age: number;
}

export interface CustomError {
  message: string;
  status: string;
}
