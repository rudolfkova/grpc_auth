-- Уникальность отображаемого имени для непустых значений (несколько записей с name = '' по-прежнему допустимы для старых данных).
CREATE UNIQUE INDEX worlds_name_unique ON worlds (name) WHERE name <> '';
