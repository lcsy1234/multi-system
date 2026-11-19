
package com.example.hello_api.dto;

public record CreateOrderRequest(
    String productCode,
    Integer quantity
) {}
