
package com.example.lianxi.service;

import com.example.lianxi.dto.ProductResponse;
import org.springframework.stereotype.Service;

@Service
public class ProductService {

    public ProductResponse getProduct(String code) {
        return new ProductResponse(
            code,
            "测试商品",
            100
        );
    }
}
