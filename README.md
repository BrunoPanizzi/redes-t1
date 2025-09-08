# T1 - Laboratório de Redes

O primeiro trabalho da discplina de laboratório de redes consiste na criação de uma aplicação cliente-servidor para a transmissão de arquivos via conexão TCP.

## O protocolo

O PRBP é um protocolo em nível de aplicação, onde uma conexão TCP é reaproveitada para a troca de comandos e respostas.

Ambas a requisição e a resposta tem o seguinte formato:

```
<protocol> <method> [payload_size]\n
[payload]
```

O campo protocol sempre deve ser `PRBP`.

Os métodos disponíveis são: 
- `LIST`: lista os arquivos armazenados no servidor;
- `PUT`: envia um arquivo para ser armazenado;
- `QUIT`: encerra a conexão.

Os campos `payload_size` e `payload` são opcionais, sendo utilizados apenas pelo método `PUT` e pela resposta do método `LIST`.

### `LIST`
Uma requisição do tipo `LIST` tem como objetivo listar os arquivos disponíveis no servidor.

Exemplo de requisição:
```
PRBP LIST\n
```

Exemplo de resposta:
```
PRBP LIST <payload_size>\n
<file_name> <file_size> <file_hash>\n
```

### `PUT`
Uma requisição do tipo `PUT` tem como objetivo enviar um arquivo para ser armazenado no servidor.

Exemplo de requisição:
```
PRBP PUT <payload_size>\n
<file_name> <file_hash>\n
<file_content>
```

Exemplo de resposta:
```
PRBP PUT <payload_size>\n
<status>
```

TODO: definir os status disponíveis para controle de erro.

### `QUIT`
Uma requisição do tipo `QUIT` encerra a conexção TCP.

Exemplo de requisição:
```
PRBP QUIT\n
```

Exemplo de resposta:
```
PRBP QUIT\n
```

TODO: definir se o servidor deve responder.

