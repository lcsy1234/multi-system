package com.example.hello_api.service;

import org.springframework.stereotype.Service;
import org.springframework.http.HttpStatus;
import org.springframework.web.server.ResponseStatusException;
import com.example.hello_api.dto.CreateOrderRequest;

@Service
public class OrderService {

    public String createOrder(CreateOrderRequest req) {

        // 校验订单数量
        if (req.quantity() == null || req.quantity() <= 0) {
            throw new ResponseStatusException(HttpStatus.BAD_REQUEST, "数量必须大于0");
        }

        // 模拟创建订单
        return "订单创建成功，商品：" + req.productCode();
    }
}
