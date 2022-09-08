INSERT INTO subcategories (id, name) VALUES (0, '');
INSERT INTO subcategories (name, category_subcategories)
VALUES
    ('клетки для кошек', (select id from categories where name = 'Клетки')),
    ('клетки для собак', (select id from categories where name = 'Клетки')),
    ('усиленная клетка (вольер)', (select id from categories where name = 'Клетки'));