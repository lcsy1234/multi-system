
package com.example.hello_api.controller;

import com.example.hello_api.dto.CreateOrderRequest;
import com.example.hello_api.service.OrderService;
import org.springframework.web.bind.annotation.*;

@RestController
@RequestMapping("/api/orders")
public class OrderController {

    private final OrderService orderService;

    public OrderController(OrderService orderService) {
        this.orderService = orderService;
    }

    @PostMapping
    public String createOrder(
        @RequestBody CreateOrderRequest req
    ) {
        return orderService.createOrder(req);
    }
}
