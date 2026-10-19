package handlers

import (
	"aiops-backend/models"
	"testing"
	"time"
)

func TestNormalizeInspectionOverviewRangeUsesStablePointIntervals(t *testing.T) {
	tests := []struct {
		value      string
		wantRange  time.Duration
		wantStep   time.Duration
		wantPoints int
	}{
		{value: "1h", wantRange: time.Hour, wantStep: 5 * time.Minute, wantPoints: 12},
		{value: "6h", wantRange: 6 * time.Hour, wantStep: 10 * time.Minute, wantPoints: 36},
		{value: "24h", wantRange: 24 * time.Hour, wantStep: 30 * time.Minute, wantPoints: 48},
		{value: "7d", wantRange: 7 * 24 * time.Hour, wantStep: 4 * time.Hour, wantPoints: 42},
		{value: "30d", wantRange: 30 * 24 * time.Hour, wantStep: 12 * time.Hour, wantPoints: 60},
	}
	for _, test := range tests {
		t.Run(test.value, func(t *testing.T) {
			_, gotRange, gotStep := normalizeInspectionOverviewRange(test.value)
			if gotRange != test.wantRange || gotStep != test.wantStep {
				t.Fatalf("range %s: got range=%s step=%s, want range=%s step=%s", test.value, gotRange, gotStep, test.wantRange, test.wantStep)
			}
			if got := int((gotRange + gotStep - 1) / gotStep); got != test.wantPoints {
				t.Fatalf("range %s: got %d points, want %d", test.value, got, test.wantPoints)
			}
		})
	}
}

func TestAggregateInspectionTrendBucketsProducesChartReadyPoints(t *testing.T) {
	start := time.Date(2026, 8, 26, 10, 0, 0, 0, time.Local)
	node := inspectionNodeView{NodeKey: "agent:7", NodeType: "agent", NodeID: 7, NodeName: "美国节点"}
	buckets := []models.InspectionTrendBucket{
		{ID: 1, BucketAt: start.Add(time.Minute), TotalCount: 2, SuccessCount: 2, DurationSumMs: 30, DurationMaxMs: 20, Status2xxCount: 2},
		{ID: 2, BucketAt: start.Add(6 * time.Minute), TotalCount: 1, FailureCount: 1, DurationSumMs: 50, DurationMaxMs: 50, Status5xxCount: 1},
		{ID: 3, BucketAt: start.Add(-time.Minute), TotalCount: 1, SuccessCount: 1, DurationSumMs: 10},
	}

	points := aggregateInspectionTrendBuckets(buckets, node, start, time.Hour, 5*time.Minute)
	if len(points) != 12 {
		t.Fatalf("got %d points, want 12", len(points))
	}
	if points[0].TotalCount != 2 || points[0].DurationMS != 15 || points[0].Status != "success" || points[0].Status2xxCount != 2 {
		t.Fatalf("unexpected first point: %+v", points[0])
	}
	if points[1].TotalCount != 1 || points[1].DurationMS != 50 || points[1].Status != "fail" || points[1].Status5xxCount != 1 {
		t.Fatalf("unexpected second point: %+v", points[1])
	}
	if !points[2].CreatedAt.Equal(start.Add(10*time.Minute)) || points[2].TotalCount != 0 {
		t.Fatalf("unexpected empty point: %+v", points[2])
	}
}
