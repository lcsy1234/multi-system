package com.example.lianxi.dto;

import jakarta.validation.constraints.NotBlank;

public record ProductQueryRequest(
    @NotBlank(message = "商品编码不能为空") String code
) {}
