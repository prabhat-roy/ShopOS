from datetime import datetime, timezone
from typing import Optional

import numpy as np

from .models import PriceOptimizationRequest, PriceOptimizationResult


class PriceOptimizer:
    """Demand-elasticity based price optimization engine."""

    NUM_SCAN_POINTS = 100

    def calculate_margin(self, price: float, cost_price: float) -> float:
        """Return gross margin as a fraction (0.0–1.0)."""
        if price <= 0:
            return 0.0
        return (price - cost_price) / price

    def estimate_demand_change(
        self, current_price: float, new_price: float, elasticity: float
    ) -> float:
        """
        Estimate percentage change in demand (as a decimal fraction).
        Uses point-elasticity: %ΔD = elasticity × %ΔP
        """
        if current_price <= 0:
            return 0.0
        pct_price_change = (new_price - current_price) / current_price
        return elasticity * pct_price_change

    def optimize(self, req: PriceOptimizationRequest) -> PriceOptimizationResult:
        """
        Find the price in [minPrice, maxPrice] that maximises revenue while
        honouring the targetMargin constraint.  Competitor prices are used to
        bias the result toward ±5 % of the average competitor price.
        """
        if req.minPrice > req.maxPrice:
            raise ValueError("minPrice must be <= maxPrice")

        # Normalise the scan range so it always includes at least two points
        lo = max(req.minPrice, req.costPrice * (1.0 + 1e-6))
        hi = req.maxPrice

        # If cost already exceeds maxPrice there is no feasible margin-positive
        # range; clamp lo to minPrice and proceed (margin will be negative but
        # we still return the best-revenue price)
        if lo > hi:
            lo = req.minPrice

        candidate_prices = np.linspace(lo, hi, self.NUM_SCAN_POINTS)

        # Baseline demand is 1.0 (normalised units)
        baseline_demand = 1.0

        best_price: float = float(req.currentPrice)
        best_revenue: float = -float("inf")
        best_demand_change: float = 0.0

        competitor_avg: Optional[float] = None
        if req.competitorPrices:
            competitor_avg = float(np.mean(req.competitorPrices))

        for p in candidate_prices:
            # Margin gate
            margin = self.calculate_margin(float(p), req.costPrice)
            if margin < req.targetMargin:
                continue

            demand_change = self.estimate_demand_change(
                req.currentPrice, float(p), req.elasticity
            )
            demand = baseline_demand * (1.0 + demand_change)
            if demand < 0:
                demand = 0.0

            revenue = float(p) * demand

            # Competitor bias: give a 3 % revenue bonus when within ±5 % of
            # average competitor price, so the scan naturally prefers
            # competitive prices when revenue is otherwise equal.
            if competitor_avg is not None and competitor_avg > 0:
                deviation = abs(float(p) - competitor_avg) / competitor_avg
                if deviation <= 0.05:
                    revenue *= 1.03

            if revenue > best_revenue:
                best_revenue = revenue
                best_price = float(p)
                best_demand_change = demand_change

        # If no feasible price was found (all below targetMargin), fall back to
        # the minimum price that still satisfies the margin target, or just
        # minPrice if that is also infeasible.
        if best_revenue == -float("inf"):
            fallback = req.costPrice / (1.0 - req.targetMargin)
            best_price = float(np.clip(fallback, req.minPrice, req.maxPrice))
            best_demand_change = self.estimate_demand_change(
                req.currentPrice, best_price, req.elasticity
            )

        # Revenue-change relative to current revenue (current demand = 1.0)
        current_revenue = req.currentPrice * baseline_demand
        final_demand = baseline_demand * (1.0 + best_demand_change)
        if final_demand < 0:
            final_demand = 0.0
        new_revenue = best_price * final_demand
        revenue_change = (
            (new_revenue - current_revenue) / current_revenue
            if current_revenue > 0
            else 0.0
        )

        final_margin = self.calculate_margin(best_price, req.costPrice)

        # Confidence: penalise large price movements and missing competitor data
        price_move = abs(best_price - req.currentPrice) / max(req.currentPrice, 1e-9)
        competitor_factor = 0.1 if req.competitorPrices else 0.0
        confidence = float(
            np.clip(1.0 - 0.5 * price_move + competitor_factor, 0.0, 1.0)
        )

        reasoning_parts: list[str] = []
        if best_price > req.currentPrice:
            reasoning_parts.append(
                f"Price increased by {(best_price - req.currentPrice):.2f} "
                f"({price_move * 100:.1f}%) to improve revenue."
            )
        elif best_price < req.currentPrice:
            reasoning_parts.append(
                f"Price reduced by {(req.currentPrice - best_price):.2f} "
                f"({price_move * 100:.1f}%) to stimulate demand."
            )
        else:
            reasoning_parts.append("Current price is already optimal.")

        reasoning_parts.append(
            f"Elasticity of {req.elasticity:.2f} implies "
            f"{best_demand_change * 100:.1f}% demand change."
        )

        if competitor_avg is not None:
            gap = best_price - competitor_avg
            reasoning_parts.append(
                f"Suggested price is {abs(gap):.2f} "
                f"{'above' if gap >= 0 else 'below'} competitor average of "
                f"{competitor_avg:.2f}."
            )

        reasoning_parts.append(
            f"Margin at suggested price: {final_margin * 100:.1f}% "
            f"(target: {req.targetMargin * 100:.1f}%)."
        )
        reasoning_parts.append(
            f"Expected revenue change: {revenue_change * 100:.1f}%."
        )

        return PriceOptimizationResult(
            productId=req.productId,
            currentPrice=req.currentPrice,
            suggestedPrice=round(best_price, 4),
            expectedDemandChange=round(best_demand_change, 6),
            expectedRevenueChange=round(revenue_change, 6),
            marginAtSuggestedPrice=round(final_margin, 6),
            confidence=round(confidence, 4),
            reasoning=" ".join(reasoning_parts),
            generatedAt=datetime.now(timezone.utc),
        )
