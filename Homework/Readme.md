<p> Curl запрросы: </p>
<p> curl localhost:9000/pickupPoint -X POST -H "Content-Type: application/json" -d '{"name":"пвз1","address":"спб","contact_details":"mail"}'</p>
<p> curl localhost:9000/pickupPoints -X GET </p>
<p> curl localhost:9000/pickupPoint/1 -X GET </p>
<p> curl localhost:9000/pickupPoint/1 -X PUT -H "Content-Type: application/json" -d '{"name":"пвз1","address":"мск","contact_details":"mail"}'</p>
<p> curl localhost:9000/pickupPoint/1 -X DELETE </p>
<p> Для использования Makefile со своим файлом конфига, в сетапе мейкфайла нужно 
изменить значения user, password, dbname, host, port, sslmode на значения из кастамного
файла конфига. </p>