let buyerProducts =
  "4bb66fdd fc5de8c5 e835860f c3ea0a90 c3ea0a90 eff231b9 86a98679 e1dbc2dc 1fa5e1cc 86a98679 bae389d5 e1dbc2dc 717b4e68 be034610 2f15ff16 b148e2fc 717b4e68 4d2befae f7898eb9 4ab4269e 3a5e73e8 c74008ac 836a1a0b 8324b3dc 44cdf60b ea6845d4 537bfb52 3485da63 6f1da24c def3c5e4 e835860f 9c56353a 90a1e574 22d225aa 41cabd38 bae389d5 def3c5e4 693fac8 1fa5e1cc 54eb94b4 ddcf5e1b 693fac8 c74008ac 9c56353a".split(
    " "
  );

let recommended = [
  {
    ProductId: "cd3de2cc",
    Name: "Fully cooked ready pasta",
    Date: "2020-08-17T00:00:00Z",
    Price: "5449",
  },
  {
    ProductId: "44b04768",
    Name: "Fat Free Refried Beans",
    Date: "2020-08-17T00:00:00Z",
    Price: "3741",
  },
  {
    ProductId: "e70d94f9",
    Name: "Vegan noodle",
    Date: "2020-08-17T00:00:00Z",
    Price: "7192",
  },
  {
    ProductId: "52a8c80a",
    Name: "Beef pot roast with gravy",
    Date: "2020-08-17T00:00:00Z",
    Price: "3463",
  },
  {
    ProductId: "b58d28f4",
    Name: "Speedy mac",
    Date: "2020-08-17T00:00:00Z",
    Price: "7355",
  },
  {
    ProductId: "262f8a66",
    Name: "Dehydrated soup greens",
    Date: "2020-08-17T00:00:00Z",
    Price: "9949",
  },
  {
    ProductId: "7c8bf4b4",
    Name: "Seasoned white meat pulled chicken with bbq sauce",
    Date: "2020-08-17T00:00:00Z",
    Price: "7171",
  },
  {
    ProductId: "d60ff146",
    Name: "Progresso Vegetable Classics Lentil Soup",
    Date: "2020-08-17T00:00:00Z",
    Price: "9407",
  },
  {
    ProductId: "1aeea6ef",
    Name: "Steamed dumplings chicken & vegetable",
    Date: "2020-08-17T00:00:00Z",
    Price: "8299",
  },
  {
    ProductId: "371bf2f8",
    Name: "Hot Dog Chili Sauce",
    Date: "2020-08-17T00:00:00Z",
    Price: "8786",
  },
].map((product) => product.ProductId);

function contains() {
  let result = false;
  
  buyerProducts.forEach((id) => {
    if (recommended.includes(id)) {
      result = true;
      return;
    }
  });

  return result;
}

console.log(contains());
