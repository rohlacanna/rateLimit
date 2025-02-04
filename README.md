# Rate Limiting em Go

Este projeto demonstra a implementação de rate limiting (limitação de taxa) em Go, utilizando a biblioteca `golang.org/x/time/rate`. O sistema limita o número de requisições por IP em um servidor HTTP.

## Funcionalidades

- Limitação de requisições por endereço IP
- Servidor HTTP básico
- Monitoramento de requisições bem-sucedidas e rejeitadas
- Relatório detalhado de resultados por IP

## Pré-requisitos

- Go 1.22 ou superior
- golang.org/x/time v0.10.0

## Instalação

1. Clone o repositório:

## Como Usar

1. Inicie o servidor:

```bash
go run main.go
```

2. O servidor estará disponível em `http://localhost:8080`

3. Para testar o rate limiting:
   - Faça múltiplas requisições para o endpoint
   - Observe os logs no console mostrando as requisições aceitas/rejeitadas
   - Verifique o relatório final com estatísticas por IP

## Configuração

O rate limiting pode ser ajustado através das seguintes variáveis:
- `rateLimitRequests`: Define o número máximo de requisições permitidas
- `requestsPerIP`: Configura o número de requisições por IP

## Licença

É de graça! 🎉 Use à vontade (MIT License)

## Autor

Feito com ☕ e código por Rômulo Silva

---
*"O código é poesia, mas às vezes parece mais um rap de protesto"* 😄
