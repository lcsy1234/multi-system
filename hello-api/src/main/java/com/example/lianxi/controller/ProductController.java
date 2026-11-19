package com.example.lianxi.controller;

import com.example.lianxi.dto.ProductQueryRequest;
import com.example.lianxi.dto.ProductResponse;
import com.example.lianxi.service.ProductService;
import jakarta.validation.Valid;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.ModelAttribute;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

@RestController
@RequestMapping("/api/products")
public class ProductController {

    private final ProductService productService;

    public ProductController(ProductService productService) {
        this.productService = productService;
    }

    @GetMapping
    public ProductResponse queryProduct(@Valid @ModelAttribute ProductQueryRequest request) {
        return productService.getProduct(request.code());
    }
}
