# Pizza shop

To run this thing first create a ".env" file.
Template:
`
DB_USER=root
DB_PASS=mypassword
DB_PEPPER=some_random_string
`



## TODO:

- [x] my sql setup
- [x] Login/register
- [ ] Accounts
    - [ ] Add a view account button (go to /account)
    - [ ] Add a log out button
    - [ ] Admin login screen
- [x] Menu
    - [x] Menu HTML prototype
    - [x] Menu CRUD (pizza, ingredient, etc)
    - [x] Display menu on client
    - [x] Display vegan/vegetarian
    - [x] Calculate price
- [ ] Shopping cart
    - [ ] Cart HTML
    - [ ] Functionality to add pizza to cart
- [ ] Ordering
    - [ ] Order CRUD
    - [ ] Ability to confirm order
    - [ ] Order confirmation screen
    - [ ] Transactions, rollback
- [ ] Delivery person
    - [ ] Delivery person CRUD
    - [ ] Delivery person login
    - [ ] Delivery person view assigned order
    - [ ] Delivery person mark order as delivered/failed/etc
- [ ] Desserts & drinks
    - [ ] Extra item CRUD
    - [ ] Add to menu
    - [ ] Ordering desserts and drinks
- [ ] Discount codes
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
- [ ] Add updating account details