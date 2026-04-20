"""Tests for the PriceOptimizer engine — 10 tests total."""
import pytest

from optimizer.engine import PriceOptimizer
from optimizer.models import PriceOptimizationRequest


@pytest.fixture
def optimizer() -> PriceOptimizer:
    return PriceOptimizer()


def _req(**kwargs) -> PriceOptimizationRequest:
    defaults = dict(
        productId="prod-001",
        currentPrice=100.0,
        costPrice=50.0,
        minPrice=60.0,
        maxPrice=150.0,
        competitorPrices=[],
        targetMargin=0.3,
        elasticity=-1.5,
    )
    defaults.update(kwargs)
    return PriceOptimizationRequest(**defaults)


# 1. Basic optimization returns a result with the correct productId
def test_basic_optimization_returns_result(optimizer):
    result = optimizer.optimize(_req())
    assert result.productId == "prod-001"
    assert result.suggestedPrice > 0


# 2. Suggested price always respects minPrice and maxPrice bounds
def test_suggested_price_within_bounds(optimizer):
    req = _req(minPrice=80.0, maxPrice=120.0)
    result = optimizer.optimize(req)
    assert req.minPrice <= result.suggestedPrice <= req.maxPrice


# 3. Margin constraint is respected — margin at suggested price >= targetMargin
def test_margin_constraint_respected(optimizer):
    req = _req(costPrice=50.0, targetMargin=0.4)
    result = optimizer.optimize(req)
    actual_margin = optimizer.calculate_margin(result.suggestedPrice, req.costPrice)
    assert actual_margin >= req.targetMargin - 1e-6


# 4. Competitor pricing influence — result skews toward competitor average
def test_competitor_pricing_influence(optimizer):
    req_no_comp = _req(competitorPrices=[])
    req_comp = _req(competitorPrices=[95.0, 97.0, 93.0])
    r_no = optimizer.optimize(req_no_comp)
    r_with = optimizer.optimize(req_comp)
    # When competitor avg is ~95, suggested price should be close to that range
    comp_avg = 95.0
    assert abs(r_with.suggestedPrice - comp_avg) <= abs(r_no.suggestedPrice - comp_avg) + 10.0


# 5. Elasticity calculation: negative elasticity → price increase reduces demand
def test_elasticity_calculation(optimizer):
    # 10% price increase with elasticity=-1.5 → -15% demand change
    change = optimizer.estimate_demand_change(100.0, 110.0, -1.5)
    assert abs(change - (-0.15)) < 1e-9


# 6. Revenue maximization — higher revenue at suggested vs current
def test_revenue_maximization(optimizer):
    req = _req(currentPrice=100.0, costPrice=30.0, minPrice=60.0, maxPrice=200.0, targetMargin=0.2)
    result = optimizer.optimize(req)
    current_revenue = req.currentPrice * 1.0
    suggested_demand = 1.0 + optimizer.estimate_demand_change(
        req.currentPrice, result.suggestedPrice, req.elasticity
    )
    suggested_revenue = result.suggestedPrice * max(suggested_demand, 0.0)
    assert suggested_revenue >= current_revenue - 1e-6


# 7. Boundary: minPrice == maxPrice → suggested price equals that value
def test_min_equals_max_price(optimizer):
    req = _req(minPrice=100.0, maxPrice=100.0, costPrice=60.0, currentPrice=100.0)
    result = optimizer.optimize(req)
    assert abs(result.suggestedPrice - 100.0) < 1e-4


# 8. Very high cost price (costPrice > maxPrice) falls back gracefully
def test_high_cost_price_fallback(optimizer):
    req = _req(costPrice=200.0, minPrice=10.0, maxPrice=150.0, targetMargin=0.3)
    result = optimizer.optimize(req)
    assert result.suggestedPrice >= req.minPrice
    assert result.suggestedPrice <= req.maxPrice


# 9. Reasoning string is non-empty
def test_reasoning_string_nonempty(optimizer):
    result = optimizer.optimize(_req())
    assert isinstance(result.reasoning, str)
    assert len(result.reasoning.strip()) > 0


# 10. Confidence is in [0.0, 1.0]
def test_confidence_range(optimizer):
    result = optimizer.optimize(_req())
    assert 0.0 <= result.confidence <= 1.0


# Bonus: calculate_margin correctness
def test_calculate_margin(optimizer):
    margin = optimizer.calculate_margin(100.0, 60.0)
    assert abs(margin - 0.4) < 1e-9


def test_calculate_margin_zero_price(optimizer):
    assert optimizer.calculate_margin(0.0, 50.0) == 0.0
