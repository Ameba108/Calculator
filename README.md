# Распределённый вычислитель арифметических выражений
Как это работает?

Оркестратор принимает на вход выражение (например 2+2 * 2) и разбивает его на подзадачи (то есть 2 + результат умножения и 2*2). Агент в это время "просит" от оркестратора задачу и когда получает ее, выдает результат и снова просит задачу.
В дирректории calculator находится сама функция калькулятора, благодаря которой происходят все расчеты.

Установка:
1. `git clone https://github.com/Ameba108/Calculator`
2. `cd Calculator`

Запуск:

Чтобы запустить калькулятор, нужно запустить оркестратор, агент и сам калькулятор. 
1. Запуск оркестратора:
   
   `cd orchestrator`
   
   `go run .`
3. Запуск агента:
   
   `cd agent`
   
   `go run .`
5. Запуск калькулятора:
   
   `cd internal/calculator`
   
   `go run .`

# Примеры 
Вот пример post-запроса (лучше всего запросы делать через postman):

`curl -v -X POST -H "Content-Type: application/json" -d "{\"expression\": \"2+2*2\"}" http://localhost:8080/api/v1/calculate`

![Скриншот-05-03-2025 19_20_06](https://github.com/user-attachments/assets/331b1d71-f145-4578-acb1-0ad0f842c154)

В ответе выдается id выражения. Чтобы получить решение нужно скопировать это id и вставить в get-запрос:

`curl http://localhost:8080/api/v1/expressions/1741202363010826900`

Получаем ответ на запрос с id выражением, самим выражением, статусом готовности и ответ.

![Скриншот-05-03-2025 19_23_08](https://github.com/user-attachments/assets/e8aa6e45-523d-49fd-a95f-107208a4c623)

# Что было реализовано из тз?
Был успешно реализован калькулятор, с оркестратором и агентом. Оркестратор может разделять выражение на задачи, а агент в свою очередь выполняет эти задачи.

# Что не было реализовано? 
1. Тесты
   Отсутсвие тестов у оркестратора, агента и функции калькулятора
2. Обработка ошибок в случае некорректных данных и т.п.
   В случае, если пользователь отправит выражение 1+x, 2/0 или что-то подобное, в ответ он получит это:
    
![Скриншот-05-03-2025 19_29_04](https://github.com/user-attachments/assets/869615cd-3f58-4a29-87e7-c12614781b54)


