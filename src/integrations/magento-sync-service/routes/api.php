<?php

use Illuminate\Support\Facades\Route;

Route::get('/healthz', fn() => response()->json(['status' => 'ok']));

Route::prefix('magento')->group(function () {
    Route::post('/sync/products', fn() => response()->json(['status' => 'sync_queued']));
    Route::post('/sync/orders', fn() => response()->json(['status' => 'sync_queued']));
});
