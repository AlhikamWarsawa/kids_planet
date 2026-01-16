INSERT INTO education_categories (name, icon, color)
VALUES
    ('Math', 'calculator', '#4CAF50'),
    ('Reading', 'book', '#2196F3'),
    ('Logic', 'brain', '#9C27B0'),
    ('Memory', 'memory', '#FF9800'),
    ('Creativity', 'palette', '#E91E63')
ON CONFLICT (name) DO NOTHING;
