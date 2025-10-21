# Pizza shop

To run this thing first create a `.env` file:

```
DB_USER=root
DB_PASS=mypassword
DB_PEPPER=some_random_string
# set this to true to reset the database each time when starting server
# DB_RESET=1
```

Run the shit:

```
go build backend/main.go 
./main
```

## TODO:

- [x] my sql setup
- [x] Login/register
- [x] Accounts
    - [x] Add a view account button (go to /account)
    - [x] Add a log out button
    - [x] Admin login screen
- [x] Menu
    - [x] Menu HTML prototype
    - [x] Menu CRUD (pizza, ingredient, etc)
    - [x] Display menu on client
    - [x] Display vegan/vegetarian
    - [x] Calculate price
- [x] Shopping cart
    - [x] Cart HTML
    - [x] Functionality to add pizza to cart
- [ ] Ordering <-A
    - [x] Order CRUD
    - [x] Ability to confirm order
    - [x] Order confirmation screen
    - [x] Transactions, rollback
- [x] Delivery person <-K
    - [x] Delivery person CRUD
    - [x] Delivery person login
    - [ ] Delivery person view assigned order
    - [ ] Delivery person mark order as delivered/failed/etc
- [x] Desserts & drinks <-A
    - [x] Extra item CRUD
    - [x] Add to menu
    - [x] Ordering desserts and drinks
- [ ] Discount codes <-K
    - [ ] Discount code CRUD
    - [ ] Admin create discount codes
    - [ ] Birthday discount code
    - [ ] Applying discount code in order
- [ ] Testing
    - [ ] Code for generating sample orders and sample accounts
- [ ] Reports
    - [ ] Undelivered orders
    - [ ] Top 3 pizzas sold in the past month
    - [ ] Earning reports filtered by X



## Optional TODO
- [ ] Add a delete account button
- [x] Add updating account details
