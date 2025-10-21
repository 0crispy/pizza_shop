# Pizza Shop - Entity Relationship Diagram (ERD)

## Tables and Relationships

```
┌─────────────────┐
│     USER        │
├─────────────────┤
│ id (PK)         │
│ username        │
│ password_hash   │
│ salt            │
│ role            │◄────┐
└─────────────────┘     │
         △              │
         │              │
    ┌────┴────┬─────────┘
    │         │
┌───┴──────┐  │  ┌──────────────┐
│ CUSTOMER │  │  │ DELIVERY_    │
├──────────┤  │  │ PERSON       │
│ id (PK)  │  │  ├──────────────┤
│ user_id◄─┘  │  │ id (PK)      │
│ name     │  │  │ user_id◄─────┘
│ gender   │  │  │ name         │
│birth_date│  │  │ vehicle_type │
│ address  │  │  │ order_id (FK)│◄──┐
│postal_code│ │  └──────────────┘   │
│pizza_cnt │  │                     │
└────┬─────┘  │                     │
     │        │                     │
     │        │  ┌──────────────────┴──┐
     │        │  │      ORDERS         │
     │        │  ├─────────────────────┤
     │        │  │ id (PK)             │
     │        │  │ customer_id (FK)◄───┘
     └────────┼─►│ timestamp           │
              │  │ status              │
              │  │ postal_code         │
              │  │ delivery_address    │
              │  │ discount_code_id(FK)│◄────┐
              │  │ delivery_person_id  │     │
              │  └──────┬──────────────┘     │
              │         │                    │
              │    ┌────┴────┐               │
              │    │         │               │
       ┌──────┴────▼┐   ┌───▼──────────┐    │
       │ORDER_PIZZA  │   │ORDER_EXTRA_  │    │
       ├─────────────┤   │ITEM          │    │
       │ id (PK)     │   ├──────────────┤    │
       │ order_id(FK)│   │ id (PK)      │    │
       │ pizza_id(FK)│   │ order_id (FK)│    │
       │ quantity    │   │ extra_item_  │    │
       └──────┬──────┘   │ id (FK)      │    │
              │          │ quantity     │    │
              │          └──────┬───────┘    │
              │                 │            │
       ┌──────▼──────┐   ┌──────▼──────┐    │
       │    PIZZA    │   │ EXTRA_ITEM  │    │
       ├─────────────┤   ├─────────────┤    │
       │ id (PK)     │   │ id (PK)     │    │
       │ name        │   │ name        │    │
       └──────┬──────┘   │ category    │    │
              │          │ price       │    │
              │          └─────────────┘    │
       ┌──────▼──────────┐                  │
       │ PIZZA_          │                  │
       │ INGREDIENT      │                  │
       ├─────────────────┤                  │
       │ pizza_id (PK,FK)│                  │
       │ ingredient_id   │                  │
       │ (PK,FK)         │                  │
       └──────┬──────────┘                  │
              │                             │
       ┌──────▼──────┐                      │
       │ INGREDIENT  │                      │
       ├─────────────┤                      │
       │ id (PK)     │                      │
       │ name        │                      │
       │ cost        │                      │
       │ has_meat    │                      │
       │ has_animal_ │                      │
       │ products    │                      │
       └─────────────┘                      │
                                            │
       ┌────────────────────────────────────┘
       │ DISCOUNT_CODE │
       ├───────────────┤
       │ id (PK)       │
       │ code          │
       │ discount_%    │
       │ is_active     │
       └───────────────┘
```

## Cardinality

- **User → Customer**: 1:1 (One user can be one customer)
- **User → Delivery Person**: 1:1 (One user can be one delivery person)
- **Customer → Orders**: 1:N (One customer can have many orders)
- **Orders → Order_Pizza**: 1:N (One order can have many pizzas)
- **Orders → Order_Extra_Item**: 1:N (One order can have many extra items)
- **Pizza → Pizza_Ingredient**: 1:N (One pizza has many ingredients)
- **Ingredient → Pizza_Ingredient**: 1:N (One ingredient can be in many pizzas)
- **Pizza ↔ Ingredient**: M:N via Pizza_Ingredient (Many-to-Many)
- **Discount_Code → Orders**: 1:N (One code can be used in many orders)
- **Delivery_Person → Orders**: 1:N (One driver can deliver many orders)

## Key Business Rules

1. **Dynamic Pricing**: Pizza price = SUM(ingredient.cost) × 1.4 (margin) × 1.09 (VAT)
2. **Dietary Classification**: 
   - Vegan = No ingredient has meat OR animal products
   - Vegetarian = No ingredient has meat (but may have animal products)
3. **Order Transaction**: All order items inserted atomically (rollback on failure)
4. **Discount**: Applied once per order, reduces total by percentage
5. **Delivery Assignment**: Each order optionally assigned to one delivery person

## Constraints

- `ingredient.cost > 0` (CHECK)
- `order_pizza.quantity > 0` (CHECK)
- `order_extra_item.quantity > 0` (CHECK)
- `extra_item.price >= 0` (CHECK)
- `discount_code.discount_percentage BETWEEN 1 AND 100` (CHECK)
- `discount_code.code` (UNIQUE)
- `ingredient.name` (UNIQUE)
- `pizza.name` (UNIQUE)
- `user.username` (UNIQUE)

## Indexes (Recommended for Performance)

```sql
-- For order lookups
CREATE INDEX idx_orders_customer ON orders(customer_id);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_timestamp ON orders(timestamp);

-- For pizza lookups
CREATE INDEX idx_pizza_ingredient ON pizza_ingredient(pizza_id);
CREATE INDEX idx_ingredient_pizza ON pizza_ingredient(ingredient_id);

-- For order items
CREATE INDEX idx_order_pizza ON order_pizza(order_id);
CREATE INDEX idx_order_extra ON order_extra_item(order_id);
```

## Sample Queries

### Get Pizza with Price
```sql
SELECT 
    p.name,
    SUM(i.cost) * 1.4 * 1.09 as price
FROM pizza p
JOIN pizza_ingredient pi ON p.id = pi.pizza_id
JOIN ingredient i ON pi.ingredient_id = i.id
GROUP BY p.id, p.name;
```

### Get Vegan Pizzas
```sql
SELECT DISTINCT p.name
FROM pizza p
WHERE p.id NOT IN (
    SELECT DISTINCT pi.pizza_id
    FROM pizza_ingredient pi
    JOIN ingredient i ON pi.ingredient_id = i.id
    WHERE i.has_meat = TRUE OR i.has_animal_products = TRUE
);
```

### Top Pizzas Last 30 Days
```sql
SELECT 
    p.name,
    SUM(op.quantity) as total_sold
FROM order_pizza op
JOIN pizza p ON op.pizza_id = p.id
JOIN orders o ON op.order_id = o.id
WHERE o.timestamp >= DATE_SUB(NOW(), INTERVAL 30 DAY)
GROUP BY p.id, p.name
ORDER BY total_sold DESC
LIMIT 3;
```

### Undelivered Orders
```sql
SELECT 
    o.id,
    c.name as customer,
    o.delivery_address,
    o.timestamp
FROM orders o
JOIN customer c ON o.customer_id = c.id
WHERE o.status = 'IN_PROGRESS'
ORDER BY o.timestamp DESC;
```
