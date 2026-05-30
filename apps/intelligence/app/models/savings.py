from datetime import datetime
from pydantic import BaseModel
from typing import List, Optional

class Milestone(BaseModel):
    date: datetime
    target_amount: float
    description: str

class SavingsPlan(BaseModel):
    user_id: str
    vault_id: str
    goal_amount: float
    current_balance: float
    start_date: datetime
    target_date: datetime
    status: str  # "on_track", "behind_schedule", "ahead_of_schedule"
    next_milestone: Milestone | None

class SavingsPlanRequest(BaseModel):
    goal_usdc: float
    time_horizon_months: int
    max_monthly_contribution_usdc: float
    vault_id: Optional[str] = None

class ScheduleEntry(BaseModel):
    month: int
    deposit: float
    expected_balance: float
    yield_earned: float

class MilestoneProjection(BaseModel):
    month: int
    expected_balance: float

class SavingsPlanResponse(BaseModel):
    achievable: bool
    required_monthly_deposit: float
    monthly_schedule: List[ScheduleEntry]
    total_yield_earned: float
    narrative: str
    milestones: List[MilestoneProjection]
