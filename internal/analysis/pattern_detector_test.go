package analysis

import (
	"strings"
	"testing"
)

func TestArchitecturePatternDetector_DetectReact(t *testing.T) {
	ci := NewComponentIdentifier()
	dfa := NewDataFlowAnalyzer(ci)
	apd := NewArchitecturePatternDetector(ci, dfa)

	tests := []struct {
		name             string
		filePath         string
		content          string
		packageJSON      string
		expectedMinConf  float64
		expectedFramework string
	}{
		{
			name:     "React with JSX and Hooks",
			filePath: "/src/components/App.tsx",
			content: `
import React, { useState, useEffect } from 'react';

const App = () => {
  const [count, setCount] = useState(0);
  
  useEffect(() => {
    document.title = 'Count: ' + count;
  }, [count]);

  return (
    <div>
      <h1>Count: {count}</h1>
      <button onClick={() => setCount(count + 1)}>Increment</button>
    </div>
  );
};

export default App;
			`,
			packageJSON:       `{"dependencies": {"react": "^18.0.0", "react-dom": "^18.0.0"}}`,
			expectedMinConf:   0.8,
			expectedFramework: "react",
		},
		{
			name:     "Next.js with SSR",
			filePath: "/pages/index.tsx",
			content: `
import { GetServerSideProps } from 'next';
import Link from 'next/link';

const HomePage = ({ data }) => {
  return (
    <div>
      <h1>Welcome to Next.js</h1>
      <Link href="/about">About</Link>
      <p>{data.message}</p>
    </div>
  );
};

export const getServerSideProps: GetServerSideProps = async () => {
  return {
    props: {
      data: { message: 'Server-side rendered' }
    }
  };
};

export default HomePage;
			`,
			packageJSON:       `{"dependencies": {"next": "^13.0.0", "react": "^18.0.0"}}`,
			expectedMinConf:   0.6,
			expectedFramework: "nextjs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := apd.detectFrameworks(tt.filePath, tt.content, tt.packageJSON)
			if err != nil {
				t.Fatalf("detectFrameworks() error = %v", err)
			}

			frameworks := apd.GetFrameworks()
			found := false
			for _, framework := range frameworks {
				if framework.Name == tt.expectedFramework {
					found = true
					if framework.Confidence < tt.expectedMinConf {
						t.Errorf("Expected confidence >= %.2f, got %.2f", tt.expectedMinConf, framework.Confidence)
					}
					if len(framework.Evidence) == 0 {
						t.Error("Expected evidence to be provided")
					}
					break
				}
			}

			if !found {
				t.Errorf("Expected to detect %s framework", tt.expectedFramework)
			}
		})
	}
}

func TestArchitecturePatternDetector_DetectVue(t *testing.T) {
	ci := NewComponentIdentifier()
	dfa := NewDataFlowAnalyzer(ci)
	apd := NewArchitecturePatternDetector(ci, dfa)

	content := `
<template>
  <div class="user-profile">
    <h1>{{ user.name }}</h1>
    <button @click="updateProfile">Update</button>
  </div>
</template>

<script>
import { ref, reactive, onMounted } from 'vue';

export default {
  name: 'UserProfile',
  setup() {
    const user = reactive({
      name: 'John Doe',
      email: 'john@example.com'
    });
    
    const loading = ref(false);
    
    const updateProfile = () => {
      loading.value = true;
      // Update logic
    };
    
    onMounted(() => {
      // Fetch user data
    });
    
    return {
      user,
      loading,
      updateProfile
    };
  }
};
</script>
	`

	packageJSON := `{"dependencies": {"vue": "^3.2.0", "@vue/composition-api": "^1.0.0"}}`

	err := apd.detectFrameworks("/src/components/UserProfile.vue", content, packageJSON)
	if err != nil {
		t.Fatalf("detectFrameworks() error = %v", err)
	}

	frameworks := apd.GetFrameworks()
	found := false
	for _, framework := range frameworks {
		if framework.Name == "vue" {
			found = true
			if framework.Confidence < 0.5 {
				t.Errorf("Expected Vue confidence >= 0.5, got %.2f", framework.Confidence)
			}
			// Check metadata
			if hasTemplate, exists := framework.Metadata["has_template"]; !exists || !hasTemplate.(bool) {
				t.Error("Expected Vue template detection")
			}
			if hasComposition, exists := framework.Metadata["has_composition"]; !exists || !hasComposition.(bool) {
				t.Error("Expected Vue Composition API detection")
			}
			break
		}
	}

	if !found {
		t.Error("Expected to detect Vue framework")
	}
}

func TestArchitecturePatternDetector_DetectExpress(t *testing.T) {
	ci := NewComponentIdentifier()
	dfa := NewDataFlowAnalyzer(ci)
	apd := NewArchitecturePatternDetector(ci, dfa)

	content := `
const express = require('express');
const cors = require('cors');

const app = express();

app.use(cors());
app.use(express.json());

app.get('/', (req, res) => {
  res.json({ message: 'Hello World!' });
});

app.get('/users/:id', async (req, res) => {
  try {
    const user = await getUserById(req.params.id);
    res.json(user);
  } catch (error) {
    res.status(500).json({ error: error.message });
  }
});

app.post('/users', async (req, res) => {
  try {
    const user = await createUser(req.body);
    res.status(201).json(user);
  } catch (error) {
    res.status(400).json({ error: error.message });
  }
});

app.listen(3000, () => {
  console.log('Server running on port 3000');
});
	`

	packageJSON := `{"dependencies": {"express": "^4.18.0", "cors": "^2.8.5"}}`

	err := apd.detectFrameworks("/src/server.js", content, packageJSON)
	if err != nil {
		t.Fatalf("detectFrameworks() error = %v", err)
	}

	frameworks := apd.GetFrameworks()
	found := false
	for _, framework := range frameworks {
		if framework.Name == "express" {
			found = true
			if framework.Confidence < 0.5 {
				t.Errorf("Expected Express confidence >= 0.5, got %.2f", framework.Confidence)
			}
			// Check metadata
			if hasApp, exists := framework.Metadata["has_app"]; !exists || !hasApp.(bool) {
				t.Error("Expected Express app detection")
			}
			if hasRoutes, exists := framework.Metadata["has_routes"]; !exists || !hasRoutes.(bool) {
				t.Error("Expected Express routes detection")
			}
			if hasMiddleware, exists := framework.Metadata["has_middleware"]; !exists || !hasMiddleware.(bool) {
				t.Error("Expected Express middleware detection")
			}
			break
		}
	}

	if !found {
		t.Error("Expected to detect Express framework")
	}
}

func TestArchitecturePatternDetector_DetectMVCArchitecture(t *testing.T) {
	ci := NewComponentIdentifier()
	dfa := NewDataFlowAnalyzer(ci)
	apd := NewArchitecturePatternDetector(ci, dfa)

	content := `
const UserModel = require('../models/UserModel');
const userService = require('../services/userService');

class UserController {
  async getAllUsers(req, res) {
    try {
      const users = await UserModel.findAll();
      res.render('users/index', { users });
    } catch (error) {
      res.status(500).json({ error: error.message });
    }
  }
  
  async getUserById(req, res) {
    try {
      const user = await UserModel.findById(req.params.id);
      if (!user) {
        return res.status(404).render('errors/404');
      }
      res.render('users/show', { user });
    } catch (error) {
      res.status(500).json({ error: error.message });
    }
  }
  
  async createUser(req, res) {
    try {
      const userData = req.body;
      const user = await UserModel.create(userData);
      res.redirect('/users');
    } catch (error) {
      res.status(400).render('users/new', { error: error.message });
    }
  }
}

module.exports = UserController;
	`

	err := apd.detectArchitecturalStyles("/src/controllers/UserController.js", content)
	if err != nil {
		t.Fatalf("detectArchitecturalStyles() error = %v", err)
	}

	architecturalStyles := apd.GetArchitecturalStyles()
	found := false
	for _, style := range architecturalStyles {
		if style.Name == "mvc" {
			found = true
			if style.Confidence < 0.6 {
				t.Errorf("Expected MVC confidence >= 0.6, got %.2f", style.Confidence)
			}
			// Check metadata
			if hasControllers, exists := style.Metadata["has_controllers"]; !exists || !hasControllers.(bool) {
				t.Error("Expected MVC controllers detection")
			}
			break
		}
	}

	if !found {
		t.Error("Expected to detect MVC architecture")
	}
}

func TestArchitecturePatternDetector_DetectCleanArchitecture(t *testing.T) {
	ci := NewComponentIdentifier()
	dfa := NewDataFlowAnalyzer(ci)
	apd := NewArchitecturePatternDetector(ci, dfa)

	content := `
interface UserRepository {
  findById(id: string): Promise<User>;
  create(user: User): Promise<User>;
  update(id: string, user: Partial<User>): Promise<User>;
  delete(id: string): Promise<void>;
}

class GetUserUseCase {
  constructor(private userRepository: UserRepository) {}
  
  async execute(id: string): Promise<User> {
    const user = await this.userRepository.findById(id);
    if (!user) {
      throw new Error('User not found');
    }
    return user;
  }
}

class CreateUserUseCase {
  constructor(
    private userRepository: UserRepository,
    private emailService: EmailService
  ) {}
  
  async execute(userData: CreateUserRequest): Promise<User> {
    const user = await this.userRepository.create(userData);
    await this.emailService.sendWelcomeEmail(user.email);
    return user;
  }
}

export { GetUserUseCase, CreateUserUseCase };
	`

	err := apd.detectArchitecturalStyles("/src/usecases/UserUseCases.ts", content)
	if err != nil {
		t.Fatalf("detectArchitecturalStyles() error = %v", err)
	}

	architecturalStyles := apd.GetArchitecturalStyles()
	found := false
	for _, style := range architecturalStyles {
		if style.Name == "clean" {
			found = true
			if style.Confidence < 0.6 {
				t.Errorf("Expected Clean Architecture confidence >= 0.6, got %.2f", style.Confidence)
			}
			// Check metadata
			if hasUseCases, exists := style.Metadata["has_use_cases"]; !exists || !hasUseCases.(bool) {
				t.Error("Expected Clean Architecture use cases detection")
			}
			break
		}
	}

	if !found {
		t.Error("Expected to detect Clean Architecture")
	}
}

func TestArchitecturePatternDetector_DetectComponentBasedArchitecture(t *testing.T) {
	ci := NewComponentIdentifier()
	dfa := NewDataFlowAnalyzer(ci)
	apd := NewArchitecturePatternDetector(ci, dfa)

	content := `
import React from 'react';

interface ButtonProps {
  children: React.ReactNode;
  variant?: 'primary' | 'secondary';
  onClick?: () => void;
  disabled?: boolean;
}

const Button: React.FC<ButtonProps> = ({
  children,
  variant = 'primary',
  onClick,
  disabled = false
}) => {
  return (
    <button
      className={variant === 'primary' ? 'btn-primary' : 'btn-secondary'}
      onClick={onClick}
      disabled={disabled}
    >
      {children}
    </button>
  );
};

export default Button;
	`

	err := apd.detectArchitecturalStyles("/src/components/Button.tsx", content)
	if err != nil {
		t.Fatalf("detectArchitecturalStyles() error = %v", err)
	}

	architecturalStyles := apd.GetArchitecturalStyles()
	found := false
	for _, style := range architecturalStyles {
		if style.Name == "component_based" {
			found = true
			if style.Confidence < 0.7 {
				t.Errorf("Expected Component-based confidence >= 0.7, got %.2f", style.Confidence)
			}
			// Check metadata
			if hasReusability, exists := style.Metadata["has_reusability"]; !exists || !hasReusability.(bool) {
				t.Error("Expected component reusability detection")
			}
			break
		}
	}

	if !found {
		t.Error("Expected to detect Component-based architecture")
	}
}

func TestArchitecturePatternDetector_DetectFactoryPattern(t *testing.T) {
	ci := NewComponentIdentifier()
	dfa := NewDataFlowAnalyzer(ci)
	apd := NewArchitecturePatternDetector(ci, dfa)

	content := `
interface Vehicle {
  start(): void;
  stop(): void;
}

class Car implements Vehicle {
  start() { console.log('Car started'); }
  stop() { console.log('Car stopped'); }
}

class Motorcycle implements Vehicle {
  start() { console.log('Motorcycle started'); }
  stop() { console.log('Motorcycle stopped'); }
}

class VehicleFactory {
  static createVehicle(type: string): Vehicle {
    switch (type) {
      case 'car':
        return new Car();
      case 'motorcycle':
        return new Motorcycle();
      default:
        throw new Error('Unknown vehicle type');
    }
  }
}

// Usage
const car = VehicleFactory.createVehicle('car');
const motorcycle = VehicleFactory.createVehicle('motorcycle');
	`

	err := apd.detectDesignPatterns("/src/patterns/VehicleFactory.ts", content)
	if err != nil {
		t.Fatalf("detectDesignPatterns() error = %v", err)
	}

	designPatterns := apd.GetDesignPatterns()
	found := false
	for _, pattern := range designPatterns {
		if pattern.Name == "factory" {
			found = true
			if pattern.Confidence < 0.7 {
				t.Errorf("Expected Factory pattern confidence >= 0.7, got %.2f", pattern.Confidence)
			}
			break
		}
	}

	if !found {
		t.Error("Expected to detect Factory pattern")
	}
}

func TestArchitecturePatternDetector_DetectRepositoryPattern(t *testing.T) {
	ci := NewComponentIdentifier()
	dfa := NewDataFlowAnalyzer(ci)
	apd := NewArchitecturePatternDetector(ci, dfa)

	content := `
interface Repository<T> {
  findById(id: string): Promise<T | null>;
  findAll(): Promise<T[]>;
  create(entity: T): Promise<T>;
  update(id: string, entity: Partial<T>): Promise<T>;
  delete(id: string): Promise<void>;
}

class UserRepository implements Repository<User> {
  private database: Database;
  
  constructor(database: Database) {
    this.database = database;
  }
  
  async findById(id: string): Promise<User | null> {
    const query = 'SELECT * FROM users WHERE id = ?';
    const result = await this.database.query(query, [id]);
    return result.length > 0 ? result[0] : null;
  }
  
  async findAll(): Promise<User[]> {
    const query = 'SELECT * FROM users';
    return await this.database.query(query);
  }
  
  async create(user: User): Promise<User> {
    const query = 'INSERT INTO users (name, email) VALUES (?, ?)';
    const result = await this.database.query(query, [user.name, user.email]);
    return { ...user, id: result.insertId };
  }
  
  async update(id: string, user: Partial<User>): Promise<User> {
    const query = 'UPDATE users SET name = ?, email = ? WHERE id = ?';
    await this.database.query(query, [user.name, user.email, id]);
    return await this.findById(id);
  }
  
  async delete(id: string): Promise<void> {
    const query = 'DELETE FROM users WHERE id = ?';
    await this.database.query(query, [id]);
  }
}

export { UserRepository };
	`

	err := apd.detectDesignPatterns("/src/repositories/UserRepository.ts", content)
	if err != nil {
		t.Fatalf("detectDesignPatterns() error = %v", err)
	}

	designPatterns := apd.GetDesignPatterns()
	found := false
	for _, pattern := range designPatterns {
		if pattern.Name == "repository" {
			found = true
			if pattern.Confidence < 0.7 {
				t.Errorf("Expected Repository pattern confidence >= 0.7, got %.2f", pattern.Confidence)
			}
			// Check metadata
			if hasRepository, exists := pattern.Metadata["has_repository"]; !exists || !hasRepository.(bool) {
				t.Error("Expected Repository naming detection")
			}
			if hasCrud, exists := pattern.Metadata["has_crud"]; !exists || !hasCrud.(bool) {
				t.Error("Expected CRUD operations detection")
			}
			break
		}
	}

	if !found {
		t.Error("Expected to detect Repository pattern")
	}
}

func TestArchitecturePatternDetector_DetectHOCPattern(t *testing.T) {
	ci := NewComponentIdentifier()
	dfa := NewDataFlowAnalyzer(ci)
	apd := NewArchitecturePatternDetector(ci, dfa)

	content := `
import React from 'react';

// Higher-Order Component for authentication
const withAuth = (WrappedComponent) => {
  return (props) => {
    const [isAuthenticated, setIsAuthenticated] = React.useState(false);
    const [loading, setLoading] = React.useState(true);
    
    React.useEffect(() => {
      checkAuthentication().then((authenticated) => {
        setIsAuthenticated(authenticated);
        setLoading(false);
      });
    }, []);
    
    if (loading) {
      return <div>Loading...</div>;
    }
    
    if (!isAuthenticated) {
      return <div>Please log in to access this page.</div>;
    }
    
    return <WrappedComponent {...props} />;
  };
};

// Higher-Order Component for loading state
const withLoading = (WrappedComponent) => {
  return ({ isLoading, ...props }) => {
    if (isLoading) {
      return <div>Loading...</div>;
    }
    
    return <WrappedComponent {...props} />;
  };
};

// Usage
const ProtectedDashboard = withAuth(withLoading(Dashboard));

export { withAuth, withLoading };
	`

	err := apd.detectDesignPatterns("/src/hocs/authHoc.tsx", content)
	if err != nil {
		t.Fatalf("detectDesignPatterns() error = %v", err)
	}

	designPatterns := apd.GetDesignPatterns()
	found := false
	for _, pattern := range designPatterns {
		if pattern.Name == "hoc" {
			found = true
			if pattern.Confidence < 0.7 {
				t.Errorf("Expected HOC pattern confidence >= 0.7, got %.2f", pattern.Confidence)
			}
			// Check metadata
			if hasHoc, exists := pattern.Metadata["has_hoc"]; !exists || !hasHoc.(bool) {
				t.Error("Expected HOC naming detection")
			}
			break
		}
	}

	if !found {
		t.Error("Expected to detect HOC pattern")
	}
}

func TestArchitecturePatternDetector_DetectHooksPattern(t *testing.T) {
	ci := NewComponentIdentifier()
	dfa := NewDataFlowAnalyzer(ci)
	apd := NewArchitecturePatternDetector(ci, dfa)

	content := `
import { useState, useEffect, useCallback, useMemo } from 'react';

// Custom hook for API data fetching
const useApi = (url) => {
  const [data, setData] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  
  useEffect(() => {
    const fetchData = async () => {
      try {
        setLoading(true);
        const response = await fetch(url);
        const result = await response.json();
        setData(result);
      } catch (err) {
        setError(err);
      } finally {
        setLoading(false);
      }
    };
    
    fetchData();
  }, [url]);
  
  return { data, loading, error };
};

// Custom hook for form handling
const useForm = (initialValues) => {
  const [values, setValues] = useState(initialValues);
  
  const handleChange = useCallback((name, value) => {
    setValues(prev => ({ ...prev, [name]: value }));
  }, []);
  
  const reset = useCallback(() => {
    setValues(initialValues);
  }, [initialValues]);
  
  const memoizedValues = useMemo(() => values, [values]);
  
  return {
    values: memoizedValues,
    handleChange,
    reset
  };
};

export { useApi, useForm };
	`

	err := apd.detectDesignPatterns("/src/hooks/customHooks.tsx", content)
	if err != nil {
		t.Fatalf("detectDesignPatterns() error = %v", err)
	}

	designPatterns := apd.GetDesignPatterns()
	found := false
	for _, pattern := range designPatterns {
		if pattern.Name == "hooks" {
			found = true
			if pattern.Confidence < 0.6 {
				t.Errorf("Expected Hooks pattern confidence >= 0.6, got %.2f", pattern.Confidence)
			}
			// Check metadata
			if hasHooks, exists := pattern.Metadata["has_hooks"]; !exists || !hasHooks.(bool) {
				t.Error("Expected hooks usage detection")
			}
			if hasCustomHooks, exists := pattern.Metadata["has_custom_hooks"]; !exists || !hasCustomHooks.(bool) {
				t.Error("Expected custom hooks detection")
			}
			break
		}
	}

	if !found {
		t.Error("Expected to detect Hooks pattern")
	}
}

func TestArchitecturePatternDetector_GetPrimaryFramework(t *testing.T) {
	ci := NewComponentIdentifier()
	dfa := NewDataFlowAnalyzer(ci)
	apd := NewArchitecturePatternDetector(ci, dfa)

	// Add multiple frameworks with different confidence levels
	apd.frameworks = append(apd.frameworks, DetectionResult{
		Type:       "framework",
		Name:       "react",
		Confidence: 0.9,
	})
	
	apd.frameworks = append(apd.frameworks, DetectionResult{
		Type:       "framework", 
		Name:       "express",
		Confidence: 0.7,
	})

	primary := apd.GetPrimaryFramework()
	if primary == nil {
		t.Fatal("Expected primary framework to be detected")
	}

	if primary.Name != "react" {
		t.Errorf("Expected primary framework to be 'react', got '%s'", primary.Name)
	}

	if primary.Confidence != 0.9 {
		t.Errorf("Expected primary framework confidence to be 0.9, got %.2f", primary.Confidence)
	}
}

func TestArchitecturePatternDetector_PatternComplianceAssessment(t *testing.T) {
	ci := NewComponentIdentifier()
	dfa := NewDataFlowAnalyzer(ci)
	apd := NewArchitecturePatternDetector(ci, dfa)

	// Add test patterns
	apd.frameworks = append(apd.frameworks, DetectionResult{
		Type:       "framework",
		Name:       "react",
		Confidence: 0.9,
	})
	
	apd.architectural = append(apd.architectural, DetectionResult{
		Type:       "architectural_style",
		Name:       "component_based",
		Confidence: 0.8,
	})
	
	apd.designPatterns = append(apd.designPatterns, DetectionResult{
		Type:       "design_pattern",
		Name:       "hooks",
		Confidence: 0.7,
	})

	assessment := apd.GetPatternComplianceAssessment()

	// Check primary framework
	if primaryFramework, exists := assessment["primary_framework"]; !exists || primaryFramework != "react" {
		t.Error("Expected primary framework to be 'react'")
	}

	// Check architectural styles count
	if stylesCount, exists := assessment["architectural_styles"]; !exists || stylesCount != 1 {
		t.Error("Expected 1 architectural style")
	}

	// Check design patterns count
	if patternsCount, exists := assessment["design_patterns_count"]; !exists || patternsCount != 1 {
		t.Error("Expected 1 design pattern")
	}

	// Check compliance assessments
	if compliance, exists := assessment["framework_compliance"]; !exists {
		t.Error("Expected framework compliance assessment")
	} else if compliance != "basic" && compliance != "good" {
		t.Errorf("Expected framework compliance to be 'basic' or 'good', got %s", compliance)
	}
}

func TestArchitecturePatternDetector_ExportToJSON(t *testing.T) {
	ci := NewComponentIdentifier()
	dfa := NewDataFlowAnalyzer(ci)
	apd := NewArchitecturePatternDetector(ci, dfa)

	// Add test data
	apd.frameworks = append(apd.frameworks, DetectionResult{
		Type:       "framework",
		Name:       "react",
		Confidence: 0.9,
		Evidence:   []string{"React imports", "JSX usage"},
	})

	jsonData, err := apd.ExportToJSON()
	if err != nil {
		t.Fatalf("ExportToJSON() error = %v", err)
	}

	if len(jsonData) == 0 {
		t.Error("Expected non-empty JSON export")
	}

	// Verify it contains expected keys
	jsonStr := string(jsonData)
	expectedKeys := []string{"frameworks", "architectural_styles", "design_patterns", "compliance_assessment"}

	for _, key := range expectedKeys {
		if !strings.Contains(jsonStr, key) {
			t.Errorf("Expected JSON to contain key: %s", key)
		}
	}
}

func TestArchitecturePatternDetector_ComprehensiveDetection(t *testing.T) {
	ci := NewComponentIdentifier()
	dfa := NewDataFlowAnalyzer(ci)
	apd := NewArchitecturePatternDetector(ci, dfa)

	// React component with multiple patterns
	content := `
import React, { useState, useEffect } from 'react';

// Custom hook
const useCounter = (initialValue = 0) => {
  const [count, setCount] = useState(initialValue);
  
  const increment = () => setCount(prev => prev + 1);
  const decrement = () => setCount(prev => prev - 1);
  const reset = () => setCount(initialValue);
  
  return { count, increment, decrement, reset };
};

// HOC for logging
const withLogging = (Component) => {
  return (props) => {
    useEffect(() => {
      console.log('Component mounted:', Component.name);
      return () => console.log('Component unmounted:', Component.name);
    }, []);
    
    return <Component {...props} />;
  };
};

// Component using patterns
const Counter = () => {
  const { count, increment, decrement, reset } = useCounter(0);
  
  return (
    <div>
      <h2>Count: {count}</h2>
      <button onClick={increment}>+</button>
      <button onClick={decrement}>-</button>
      <button onClick={reset}>Reset</button>
    </div>
  );
};

const EnhancedCounter = withLogging(Counter);

export default EnhancedCounter;
	`

	packageJSON := `{
		"dependencies": {
			"react": "^18.0.0",
			"react-dom": "^18.0.0"
		}
	}`

	err := apd.DetectPatterns("/src/components/Counter.tsx", content, packageJSON)
	if err != nil {
		t.Fatalf("DetectPatterns() error = %v", err)
	}

	// Should detect React framework
	frameworks := apd.GetFrameworks()
	if len(frameworks) == 0 {
		t.Error("Expected to detect at least one framework")
	}

	reactFound := false
	for _, framework := range frameworks {
		if framework.Name == "react" {
			reactFound = true
			if framework.Confidence < 0.8 {
				t.Errorf("Expected high React confidence, got %.2f", framework.Confidence)
			}
			break
		}
	}
	if !reactFound {
		t.Error("Expected to detect React framework")
	}

	// Should detect design patterns
	patterns := apd.GetDesignPatterns()
	expectedPatterns := []string{"hooks", "hoc"}
	
	for _, expectedPattern := range expectedPatterns {
		found := false
		for _, pattern := range patterns {
			if pattern.Name == expectedPattern {
				found = true
				if pattern.Confidence < 0.6 {
					t.Errorf("Expected %s pattern confidence >= 0.6, got %.2f", expectedPattern, pattern.Confidence)
				}
				break
			}
		}
		if !found {
			t.Errorf("Expected to detect %s pattern", expectedPattern)
		}
	}

	// Should have architectural style
	architecturalStyles := apd.GetArchitecturalStyles()
	componentBasedFound := false
	for _, style := range architecturalStyles {
		if style.Name == "component_based" {
			componentBasedFound = true
			break
		}
	}
	if !componentBasedFound {
		t.Error("Expected to detect component-based architecture")
	}
}