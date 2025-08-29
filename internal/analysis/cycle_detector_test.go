package analysis

import (
	"strings"
	"testing"
)

func TestCycleDetector_ExtractDependencies(t *testing.T) {
	ci := NewComponentIdentifier()
	cd := NewCycleDetector(ci)

	tests := []struct {
		name         string
		filePath     string
		content      string
		expectedDeps []string
	}{
		{
			name:     "ES6 Import Dependencies",
			filePath: "/src/components/Button.jsx",
			content: `
import React from 'react';
import { useState } from 'react';
import utils from './utils.js';
import Header from '../components/Header.jsx';
import './Button.css';
			`,
			expectedDeps: []string{"/src/components/utils.js", "/src/components/Header.jsx"},
		},
		{
			name:     "CommonJS Require Dependencies",
			filePath: "/src/services/api.js",
			content: `
const express = require('express');
const utils = require('./utils');
const config = require('../config/database');
const userService = require('./userService.js');
			`,
			expectedDeps: []string{"/src/services/utils.js", "/src/config/database.js", "/src/services/userService.js"},
		},
		{
			name:     "Mixed Import Types",
			filePath: "/src/components/App.tsx",
			content: `
import React, { Component } from 'react';
import { BrowserRouter } from 'react-router-dom';
const helper = require('./helper.ts');
import('./dynamicModule').then(module => {});
import config from '../config/app.config.js';
			`,
			expectedDeps: []string{"/src/components/helper.ts", "/src/components/dynamicModule.js", "/src/config/app.config.js"},
		},
		{
			name:     "No Relative Dependencies",
			filePath: "/src/components/External.js",
			content: `
import React from 'react';
import axios from 'axios';
import lodash from 'lodash';
const express = require('express');
			`,
			expectedDeps: []string{},
		},
		{
			name:     "Self Reference Filtered",
			filePath: "/src/utils/helper.js",
			content: `
import config from '../config/app.js';
import helper from './helper.js';
const utils = require('./helper');
			`,
			expectedDeps: []string{"/src/config/app.js"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deps, err := cd.extractDependencies(tt.filePath, tt.content)
			if err != nil {
				t.Fatalf("extractDependencies() error = %v", err)
			}

			if len(deps) != len(tt.expectedDeps) {
				t.Errorf("Expected %d dependencies, got %d: %v", len(tt.expectedDeps), len(deps), deps)
				return
			}

			for _, expectedDep := range tt.expectedDeps {
				found := false
				for _, dep := range deps {
					if dep == expectedDep {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected dependency %s not found in %v", expectedDep, deps)
				}
			}
		})
	}
}

func TestCycleDetector_DetectSimpleCycle(t *testing.T) {
	ci := NewComponentIdentifier()
	cd := NewCycleDetector(ci)

	// Create a simple A -> B -> A cycle
	fileA := "/src/components/A.js"
	contentA := `import B from './B.js';`

	fileB := "/src/components/B.js"
	contentB := `import A from './A.js';`

	// Add files to dependency graph
	err := cd.DetectCycles(fileA, contentA)
	if err != nil {
		t.Fatalf("DetectCycles() error for file A: %v", err)
	}

	err = cd.DetectCycles(fileB, contentB)
	if err != nil {
		t.Fatalf("DetectCycles() error for file B: %v", err)
	}

	// Analyze cycles
	err = cd.AnalyzeCycles()
	if err != nil {
		t.Fatalf("AnalyzeCycles() error: %v", err)
	}

	cycles := cd.GetCycles()
	if len(cycles) != 1 {
		t.Fatalf("Expected 1 cycle, got %d", len(cycles))
	}

	cycle := cycles[0]
	if cycle.Length != 2 {
		t.Errorf("Expected cycle length 2, got %d", cycle.Length)
	}

	if cycle.Type != ComponentCycle {
		t.Errorf("Expected cycle type %v, got %v", ComponentCycle, cycle.Type)
	}

	if cycle.Severity == "" {
		t.Error("Expected cycle severity to be set")
	}
}

func TestCycleDetector_DetectComplexCycle(t *testing.T) {
	ci := NewComponentIdentifier()
	cd := NewCycleDetector(ci)

	// Create a complex A -> B -> C -> A cycle
	files := map[string]string{
		"/src/components/A.js": `import B from './B.js';`,
		"/src/components/B.js": `import C from './C.js';`,
		"/src/components/C.js": `import A from './A.js';`,
	}

	// Add all files to dependency graph
	for filePath, content := range files {
		err := cd.DetectCycles(filePath, content)
		if err != nil {
			t.Fatalf("DetectCycles() error for file %s: %v", filePath, err)
		}
	}

	// Analyze cycles
	err := cd.AnalyzeCycles()
	if err != nil {
		t.Fatalf("AnalyzeCycles() error: %v", err)
	}

	cycles := cd.GetCycles()
	if len(cycles) != 1 {
		t.Fatalf("Expected 1 cycle, got %d", len(cycles))
	}

	cycle := cycles[0]
	if cycle.Length != 3 {
		t.Errorf("Expected cycle length 3, got %d", cycle.Length)
	}

	// Should be high severity for length >= 3
	if cycle.Severity != HighSeverity {
		t.Errorf("Expected high severity for 3-file cycle, got %v", cycle.Severity)
	}
}

func TestCycleDetector_DetectComponentCycle(t *testing.T) {
	ci := NewComponentIdentifier()
	cd := NewCycleDetector(ci)

	// Create a component cycle
	files := map[string]string{
		"/src/components/UserComponent.jsx": `
import React from 'react';
import ProfileComponent from './ProfileComponent.jsx';
		`,
		"/src/components/ProfileComponent.jsx": `
import React from 'react';
import UserComponent from './UserComponent.jsx';
		`,
	}

	// Add all files to dependency graph
	for filePath, content := range files {
		err := cd.DetectCycles(filePath, content)
		if err != nil {
			t.Fatalf("DetectCycles() error for file %s: %v", filePath, err)
		}
	}

	// Analyze cycles
	err := cd.AnalyzeCycles()
	if err != nil {
		t.Fatalf("AnalyzeCycles() error: %v", err)
	}

	cycles := cd.GetCycles()
	if len(cycles) != 1 {
		t.Fatalf("Expected 1 cycle, got %d", len(cycles))
	}

	cycle := cycles[0]
	if cycle.Type != ComponentCycle {
		t.Errorf("Expected component cycle type, got %v", cycle.Type)
	}

	// Component cycles should be high severity
	if cycle.Severity != HighSeverity {
		t.Errorf("Expected high severity for component cycle, got %v", cycle.Severity)
	}

	// Check that resolution strategies include dependency injection
	foundDI := false
	for _, strategy := range cycle.Resolution {
		if strings.Contains(strategy.Strategy, "Dependency Injection") {
			foundDI = true
			break
		}
	}
	if !foundDI {
		t.Error("Expected dependency injection strategy for component cycle")
	}
}

func TestCycleDetector_DetectServiceCycle(t *testing.T) {
	ci := NewComponentIdentifier()
	cd := NewCycleDetector(ci)

	// Create a service cycle
	files := map[string]string{
		"/src/services/userService.js": `
const profileService = require('./profileService.js');
class UserService {}
		`,
		"/src/services/profileService.js": `
const userService = require('./userService.js');
class ProfileService {}
		`,
	}

	// Add all files to dependency graph
	for filePath, content := range files {
		err := cd.DetectCycles(filePath, content)
		if err != nil {
			t.Fatalf("DetectCycles() error for file %s: %v", filePath, err)
		}
	}

	// Analyze cycles
	err := cd.AnalyzeCycles()
	if err != nil {
		t.Fatalf("AnalyzeCycles() error: %v", err)
	}

	cycles := cd.GetCycles()
	if len(cycles) != 1 {
		t.Fatalf("Expected 1 cycle, got %d", len(cycles))
	}

	cycle := cycles[0]
	if cycle.Type != ServiceCycle {
		t.Errorf("Expected service cycle type, got %v", cycle.Type)
	}
}

func TestCycleDetector_DetectTypeCycle(t *testing.T) {
	ci := NewComponentIdentifier()
	cd := NewCycleDetector(ci)

	// Create a type cycle
	files := map[string]string{
		"/src/types/user.d.ts": `
import { Profile } from './profile';
export interface User {
  profile: Profile;
}
		`,
		"/src/types/profile.d.ts": `
import { User } from './user';
export interface Profile {
  user: User;
}
		`,
	}

	// Add all files to dependency graph
	for filePath, content := range files {
		err := cd.DetectCycles(filePath, content)
		if err != nil {
			t.Fatalf("DetectCycles() error for file %s: %v", filePath, err)
		}
	}

	// Analyze cycles
	err := cd.AnalyzeCycles()
	if err != nil {
		t.Fatalf("AnalyzeCycles() error: %v", err)
	}

	cycles := cd.GetCycles()
	if len(cycles) != 1 {
		t.Fatalf("Expected 1 cycle, got %d", len(cycles))
	}

	cycle := cycles[0]
	if cycle.Type != TypeCycle {
		t.Errorf("Expected type cycle, got %v", cycle.Type)
	}

	// Type cycles should be critical severity
	if cycle.Severity != CriticalSeverity {
		t.Errorf("Expected critical severity for type cycle, got %v", cycle.Severity)
	}

	// Check that resolution strategies include type abstractions
	foundTypeAbstraction := false
	for _, strategy := range cycle.Resolution {
		if strings.Contains(strategy.Strategy, "Type Abstractions") {
			foundTypeAbstraction = true
			break
		}
	}
	if !foundTypeAbstraction {
		t.Error("Expected type abstraction strategy for type cycle")
	}
}

func TestCycleDetector_NoCycles(t *testing.T) {
	ci := NewComponentIdentifier()
	cd := NewCycleDetector(ci)

	// Create files with no cycles
	files := map[string]string{
		"/src/components/A.js": `import B from './B.js';`,
		"/src/components/B.js": `import C from './C.js';`,
		"/src/components/C.js": `// No imports`,
	}

	// Add all files to dependency graph
	for filePath, content := range files {
		err := cd.DetectCycles(filePath, content)
		if err != nil {
			t.Fatalf("DetectCycles() error for file %s: %v", filePath, err)
		}
	}

	// Analyze cycles
	err := cd.AnalyzeCycles()
	if err != nil {
		t.Fatalf("AnalyzeCycles() error: %v", err)
	}

	cycles := cd.GetCycles()
	if len(cycles) != 0 {
		t.Fatalf("Expected 0 cycles, got %d", len(cycles))
	}
}

func TestCycleDetector_MultipleCycles(t *testing.T) {
	ci := NewComponentIdentifier()
	cd := NewCycleDetector(ci)

	// Create multiple separate cycles
	files := map[string]string{
		// First cycle: A -> B -> A
		"/src/components/A.js": `import B from './B.js';`,
		"/src/components/B.js": `import A from './A.js';`,
		// Second cycle: C -> D -> C
		"/src/services/C.js": `import D from './D.js';`,
		"/src/services/D.js": `import C from './C.js';`,
		// No cycle: E -> F
		"/src/utils/E.js": `import F from './F.js';`,
		"/src/utils/F.js": `// No imports`,
	}

	// Add all files to dependency graph
	for filePath, content := range files {
		err := cd.DetectCycles(filePath, content)
		if err != nil {
			t.Fatalf("DetectCycles() error for file %s: %v", filePath, err)
		}
	}

	// Analyze cycles
	err := cd.AnalyzeCycles()
	if err != nil {
		t.Fatalf("AnalyzeCycles() error: %v", err)
	}

	cycles := cd.GetCycles()
	if len(cycles) != 2 {
		t.Fatalf("Expected 2 cycles, got %d", len(cycles))
	}

	// Verify both cycles are detected
	foundComponentCycle := false
	foundServiceCycle := false

	for _, cycle := range cycles {
		if cycle.Type == ComponentCycle {
			foundComponentCycle = true
		} else if cycle.Type == ServiceCycle {
			foundServiceCycle = true
		}
	}

	if !foundComponentCycle {
		t.Error("Expected to find component cycle")
	}
	if !foundServiceCycle {
		t.Error("Expected to find service cycle")
	}
}

func TestCycleDetector_GetCyclesBySeverity(t *testing.T) {
	ci := NewComponentIdentifier()
	cd := NewCycleDetector(ci)

	// Create cycles of different severities
	files := map[string]string{
		// Critical: Type cycle
		"/src/types/user.d.ts":    `import { Profile } from './profile.d.ts';`,
		"/src/types/profile.d.ts": `import { User } from './user.d.ts';`,
		// High: Component cycle (length 2)
		"/src/components/A.jsx": `import B from './B.jsx';`,
		"/src/components/B.jsx": `import A from './A.jsx';`,
	}

	// Add all files to dependency graph
	for filePath, content := range files {
		err := cd.DetectCycles(filePath, content)
		if err != nil {
			t.Fatalf("DetectCycles() error for file %s: %v", filePath, err)
		}
	}

	// Analyze cycles
	err := cd.AnalyzeCycles()
	if err != nil {
		t.Fatalf("AnalyzeCycles() error: %v", err)
	}

	// Test filtering by critical severity
	criticalCycles := cd.GetCyclesBySeverity(CriticalSeverity)
	if len(criticalCycles) != 1 {
		t.Errorf("Expected 1 critical cycle, got %d", len(criticalCycles))
	}
	if len(criticalCycles) > 0 && criticalCycles[0].Type != TypeCycle {
		t.Error("Expected critical cycle to be type cycle")
	}

	// Test filtering by high severity
	highCycles := cd.GetCyclesBySeverity(HighSeverity)
	if len(highCycles) != 1 {
		t.Errorf("Expected 1 high severity cycle, got %d", len(highCycles))
	}
	if len(highCycles) > 0 && highCycles[0].Type != ComponentCycle {
		t.Error("Expected high severity cycle to be component cycle")
	}
}

func TestCycleDetector_GetCyclesByType(t *testing.T) {
	ci := NewComponentIdentifier()
	cd := NewCycleDetector(ci)

	// Create cycles of different types
	files := map[string]string{
		// Component cycle
		"/src/components/User.jsx":    `import Profile from './Profile.jsx';`,
		"/src/components/Profile.jsx": `import User from './User.jsx';`,
		// Service cycle
		"/src/services/authService.js": `import userService from './userService.js';`,
		"/src/services/userService.js": `import authService from './authService.js';`,
	}

	// Add all files to dependency graph
	for filePath, content := range files {
		err := cd.DetectCycles(filePath, content)
		if err != nil {
			t.Fatalf("DetectCycles() error for file %s: %v", filePath, err)
		}
	}

	// Analyze cycles
	err := cd.AnalyzeCycles()
	if err != nil {
		t.Fatalf("AnalyzeCycles() error: %v", err)
	}

	// Test filtering by component type
	componentCycles := cd.GetCyclesByType(ComponentCycle)
	if len(componentCycles) != 1 {
		t.Errorf("Expected 1 component cycle, got %d", len(componentCycles))
	}

	// Test filtering by service type
	serviceCycles := cd.GetCyclesByType(ServiceCycle)
	if len(serviceCycles) != 1 {
		t.Errorf("Expected 1 service cycle, got %d", len(serviceCycles))
	}
}

func TestCycleDetector_GetCycleStats(t *testing.T) {
	ci := NewComponentIdentifier()
	cd := NewCycleDetector(ci)

	// Create various cycles for statistics
	files := map[string]string{
		// Critical type cycle
		"/src/types/user.d.ts":    `import { Profile } from './profile.d.ts';`,
		"/src/types/profile.d.ts": `import { User } from './user.d.ts';`,
		// High component cycle
		"/src/components/A.jsx": `import B from './B.jsx';`,
		"/src/components/B.jsx": `import C from './C.jsx';`,
		"/src/components/C.jsx": `import A from './A.jsx';`,
		// Medium import cycle
		"/src/utils/helper1.js": `import helper2 from './helper2.js';`,
		"/src/utils/helper2.js": `import helper1 from './helper1.js';`,
	}

	// Add all files to dependency graph
	for filePath, content := range files {
		err := cd.DetectCycles(filePath, content)
		if err != nil {
			t.Fatalf("DetectCycles() error for file %s: %v", filePath, err)
		}
	}

	// Analyze cycles
	err := cd.AnalyzeCycles()
	if err != nil {
		t.Fatalf("AnalyzeCycles() error: %v", err)
	}

	stats := cd.GetCycleStats()

	// Test total cycles
	if stats.TotalCycles != 3 {
		t.Errorf("Expected 3 total cycles, got %d", stats.TotalCycles)
	}

	// Test cycles by severity
	if stats.CyclesBySeverity[CriticalSeverity] != 1 {
		t.Errorf("Expected 1 critical cycle, got %d", stats.CyclesBySeverity[CriticalSeverity])
	}
	if stats.CyclesBySeverity[HighSeverity] != 1 {
		t.Errorf("Expected 1 high severity cycle, got %d", stats.CyclesBySeverity[HighSeverity])
	}
	if stats.CyclesBySeverity[MediumSeverity] != 1 {
		t.Errorf("Expected 1 medium severity cycle, got %d", stats.CyclesBySeverity[MediumSeverity])
	}

	// Test cycles by type
	if stats.CyclesByType[TypeCycle] != 1 {
		t.Errorf("Expected 1 type cycle, got %d", stats.CyclesByType[TypeCycle])
	}
	if stats.CyclesByType[ComponentCycle] != 1 {
		t.Errorf("Expected 1 component cycle, got %d", stats.CyclesByType[ComponentCycle])
	}

	// Test average cycle length (2 + 3 + 2) / 3 = 2.33
	expectedAverage := float64(7) / float64(3) // 2.33...
	if stats.AverageCycleLength < expectedAverage-0.1 || stats.AverageCycleLength > expectedAverage+0.1 {
		t.Errorf("Expected average cycle length ~%.2f, got %.2f", expectedAverage, stats.AverageCycleLength)
	}

	// Test cycle complexity score
	if stats.CycleComplexityScore <= 0 {
		t.Error("Expected positive cycle complexity score")
	}

	// Test most problematic files
	if len(stats.MostProblematicFiles) == 0 {
		t.Error("Expected most problematic files to be identified")
	}
}

func TestCycleDetector_ExportToJSON(t *testing.T) {
	ci := NewComponentIdentifier()
	cd := NewCycleDetector(ci)

	// Create a simple cycle
	files := map[string]string{
		"/src/components/A.js": `import B from './B.js';`,
		"/src/components/B.js": `import A from './A.js';`,
	}

	// Add files and analyze
	for filePath, content := range files {
		err := cd.DetectCycles(filePath, content)
		if err != nil {
			t.Fatalf("DetectCycles() error for file %s: %v", filePath, err)
		}
	}

	err := cd.AnalyzeCycles()
	if err != nil {
		t.Fatalf("AnalyzeCycles() error: %v", err)
	}

	// Export to JSON
	jsonStr, err := cd.ExportToJSON()
	if err != nil {
		t.Fatalf("ExportToJSON() error: %v", err)
	}

	// Basic validation that JSON contains expected fields
	if !strings.Contains(jsonStr, "cycles") {
		t.Error("Expected JSON to contain 'cycles' field")
	}
	if !strings.Contains(jsonStr, "stats") {
		t.Error("Expected JSON to contain 'stats' field")
	}
	if !strings.Contains(jsonStr, "graph") {
		t.Error("Expected JSON to contain 'graph' field")
	}
}

func TestCycleDetector_CycleNormalization(t *testing.T) {
	ci := NewComponentIdentifier()
	cd := NewCycleDetector(ci)

	// Create the same cycle but detected in different orders
	// This tests that cycles are normalized and not duplicated
	files := map[string]string{
		"/src/a.js": `import b from './b.js'; import c from './c.js';`,
		"/src/b.js": `import c from './c.js';`,
		"/src/c.js": `import a from './a.js';`,
	}

	// Add all files to dependency graph
	for filePath, content := range files {
		err := cd.DetectCycles(filePath, content)
		if err != nil {
			t.Fatalf("DetectCycles() error for file %s: %v", filePath, err)
		}
	}

	// Analyze cycles
	err := cd.AnalyzeCycles()
	if err != nil {
		t.Fatalf("AnalyzeCycles() error: %v", err)
	}

	cycles := cd.GetCycles()

	// Should detect exactly one cycle despite multiple paths
	if len(cycles) != 1 {
		t.Fatalf("Expected 1 normalized cycle, got %d", len(cycles))
	}

	cycle := cycles[0]
	if cycle.Length != 3 {
		t.Errorf("Expected cycle length 3, got %d", cycle.Length)
	}
}

func TestCycleDetector_ResolutionStrategies(t *testing.T) {
	ci := NewComponentIdentifier()

	// Test different cycle types have appropriate resolution strategies
	testCases := []struct {
		name             string
		files            map[string]string
		expectedType     CycleType
		mustHaveStrategy string
	}{
		{
			name: "Component Cycle - Should suggest DI",
			files: map[string]string{
				"/src/components/UserComponent.jsx":    `import Profile from './ProfileComponent.jsx';`,
				"/src/components/ProfileComponent.jsx": `import User from './UserComponent.jsx';`,
			},
			expectedType:     ComponentCycle,
			mustHaveStrategy: "Dependency Injection",
		},
		{
			name: "Type Cycle - Should suggest abstractions",
			files: map[string]string{
				"/src/types/user.d.ts":    `import { Profile } from './profile.d.ts';`,
				"/src/types/profile.d.ts": `import { User } from './user.d.ts';`,
			},
			expectedType:     TypeCycle,
			mustHaveStrategy: "Type Abstractions",
		},
		{
			name: "Simple Cycle - Should suggest merge",
			files: map[string]string{
				"/src/utils/a.js": `import b from './b.js';`,
				"/src/utils/b.js": `import a from './a.js';`,
			},
			expectedType:     UtilityCycle,
			mustHaveStrategy: "Merge Modules",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			detector := NewCycleDetector(ci) // Fresh detector for each test

			// Add files and analyze
			for filePath, content := range tc.files {
				err := detector.DetectCycles(filePath, content)
				if err != nil {
					t.Fatalf("DetectCycles() error: %v", err)
				}
			}

			err := detector.AnalyzeCycles()
			if err != nil {
				t.Fatalf("AnalyzeCycles() error: %v", err)
			}

			cycles := detector.GetCycles()
			if len(cycles) != 1 {
				t.Fatalf("Expected 1 cycle, got %d", len(cycles))
			}

			cycle := cycles[0]
			if cycle.Type != tc.expectedType {
				t.Errorf("Expected cycle type %v, got %v", tc.expectedType, cycle.Type)
			}

			// Check for required resolution strategy
			found := false
			for _, strategy := range cycle.Resolution {
				if strings.Contains(strategy.Strategy, tc.mustHaveStrategy) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected to find resolution strategy containing '%s'", tc.mustHaveStrategy)
			}
		})
	}
}
