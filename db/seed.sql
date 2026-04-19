INSERT INTO funds (
    name,
    vintage_year,
    target_size_usd,
    status_id
)
VALUES (
    'Titanbay Growth Fund I',
    2024,
    250000000.00,
    1
);

INSERT INTO investors (
    name,
    investor_type_id,
    email
)
VALUES (
    'Goldman Sachs Asset Management',
    2,
    'investments@gsam.com'
);

INSERT INTO investments (
    fund_id,
    investor_id,
    amount_usd,
    investment_date
)
SELECT
    f.id,
    i.id,
    50000000.00,
    CURRENT_DATE
FROM funds f, investors i
LIMIT 1;
