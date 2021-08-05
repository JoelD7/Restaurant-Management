let buyerProducts =
  "4bb66fdd fc5de8c5 e835860f c3ea0a90 c3ea0a90 eff231b9 86a98679 e1dbc2dc 1fa5e1cc 86a98679 bae389d5 e1dbc2dc 717b4e68 be034610 2f15ff16 b148e2fc 717b4e68 4d2befae f7898eb9 4ab4269e 3a5e73e8 c74008ac 836a1a0b 8324b3dc 44cdf60b ea6845d4 537bfb52 3485da63 6f1da24c def3c5e4 e835860f 9c56353a 90a1e574 22d225aa 41cabd38 bae389d5 def3c5e4 693fac8 1fa5e1cc 54eb94b4 ddcf5e1b 693fac8 c74008ac 9c56353a".split(
    " "
  );

let recommended = [
    {
        "ProductId": "990f5c21",
        "Name": "The original chicken flavor ramen noodle soup",
        "Date": "2020-08-17T00:00:00Z",
        "Price": "880"
    },
    {
        "ProductId": "79787e4a",
        "Name": "Parmesan Couscous Mix",
        "Date": "2020-08-17T00:00:00Z",
        "Price": "9813"
    },
    {
        "ProductId": "d41377ea",
        "Name": "Deli caribbean-style chicken breast salad with mango chutney and roasted red peppers",
        "Date": "2020-08-17T00:00:00Z",
        "Price": "5792"
    },
    {
        "ProductId": "cd3de2cc",
        "Name": "Fully cooked ready pasta",
        "Date": "2020-08-17T00:00:00Z",
        "Price": "5449"
    },
    {
        "ProductId": "371bf2f8",
        "Name": "Hot Dog Chili Sauce",
        "Date": "2020-08-17T00:00:00Z",
        "Price": "8786"
    },
    {
        "ProductId": "7e658b37",
        "Name": "CHEF BOYARDEE Spaghetti And Meatballs",
        "Date": "2020-08-17T00:00:00Z",
        "Price": "4262"
    },
    {
        "ProductId": "b1cb1f97",
        "Name": "Hearty homestyle corned beef hash",
        "Date": "2020-08-17T00:00:00Z",
        "Price": "9528"
    },
    {
        "ProductId": "b6c4ff46",
        "Name": "Five cheese thin & crispy crust pizza",
        "Date": "2020-08-17T00:00:00Z",
        "Price": "3781"
    },
    {
        "ProductId": "e4356fea",
        "Name": "Fully cooked cajun style turkey",
        "Date": "2020-08-17T00:00:00Z",
        "Price": "7714"
    },
    {
        "ProductId": "57633cd5",
        "Name": "Deluxe cheezy mac",
        "Date": "2020-08-17T00:00:00Z",
        "Price": "1480"
    }
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

recommended.push('990f5c21')
let set = new Set(recommended)
set.forEach((value)=>{
    console.log(value);
})


