package analysis

import (
	"encoding/json"
	"strings"
)

// FrameworkType represents different web frameworks
type FrameworkType string

const (
	ReactFramework   FrameworkType = "react"
	VueFramework     FrameworkType = "vue"
	AngularFramework FrameworkType = "angular"
	NextJSFramework  FrameworkType = "nextjs"
	ExpressFramework FrameworkType = "express"
	FastifyFramework FrameworkType = "fastify"
	NestJSFramework  FrameworkType = "nestjs"
	UnknownFramework FrameworkType = "unknown"
)

// ArchitecturalStyle represents different architectural patterns
type ArchitecturalStyle string

const (
	MVCArchitecture       ArchitecturalStyle = "mvc"
	CleanArchitecture     ArchitecturalStyle = "clean"
	HexagonalArchitecture ArchitecturalStyle = "hexagonal"
	LayeredArchitecture   ArchitecturalStyle = "layered"
	MicroservicesArch     ArchitecturalStyle = "microservices"
	MonolithicArch        ArchitecturalStyle = "monolithic"
	ComponentBasedArch    ArchitecturalStyle = "component_based"
	UnknownArchitecture   ArchitecturalStyle = "unknown"
)

// DesignPattern represents common design patterns
type DesignPattern string

const (
	FactoryPattern    DesignPattern = "factory"
	SingletonPattern  DesignPattern = "singleton"
	ObserverPattern   DesignPattern = "observer"
	StrategyPattern   DesignPattern = "strategy"
	RepositoryPattern DesignPattern = "repository"
	DecoratorPattern  DesignPattern = "decorator"
	AdapterPattern    DesignPattern = "adapter"
	MiddlewarePattern DesignPattern = "middleware"
	HOCPattern        DesignPattern = "hoc" // Higher Order Component
	HooksPattern      DesignPattern = "hooks"
)

// DetectionResult represents the result of pattern detection
type DetectionResult struct {
	Type       string                 `json:"type"`
	Name       string                 `json:"name"`
	Confidence float64                `json:"confidence"`
	Evidence   []string               `json:"evidence"`
	Location   string                 `json:"location"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// ArchitecturePatternDetector detects architectural patterns and frameworks
type ArchitecturePatternDetector struct {
	frameworks     []DetectionResult
	architectural  []DetectionResult
	designPatterns []DetectionResult
	componentId    *ComponentIdentifier
	dataFlowAn     *DataFlowAnalyzer
}

// NewArchitecturePatternDetector creates a new pattern detector instance
func NewArchitecturePatternDetector(ci *ComponentIdentifier, dfa *DataFlowAnalyzer) *ArchitecturePatternDetector {
	return &ArchitecturePatternDetector{
		frameworks:     make([]DetectionResult, 0),
		architectural:  make([]DetectionResult, 0),
		designPatterns: make([]DetectionResult, 0),
		componentId:    ci,
		dataFlowAn:     dfa,
	}
}

// DetectPatterns performs comprehensive pattern detection
func (apd *ArchitecturePatternDetector) DetectPatterns(filePath string, content string, packageJSON string) error {
	// Detect frameworks first
	if err := apd.detectFrameworks(filePath, content, packageJSON); err != nil {
		return err
	}

	// Detect architectural styles
	if err := apd.detectArchitecturalStyles(filePath, content); err != nil {
		return err
	}

	// Detect design patterns
	if err := apd.detectDesignPatterns(filePath, content); err != nil {
		return err
	}

	return nil
}

// detectFrameworks identifies the primary framework(s) being used
func (apd *ArchitecturePatternDetector) detectFrameworks(filePath, content, packageJSON string) error {
	// React detection
	reactConf := apd.detectReact(content, packageJSON)
	if reactConf > 0.5 {
		result := DetectionResult{
			Type:       "framework",
			Name:       string(ReactFramework),
			Confidence: reactConf,
			Evidence:   apd.getReactEvidence(content, packageJSON),
			Location:   filePath,
			Metadata: map[string]interface{}{
				"has_jsx":        strings.Contains(content, "jsx") || strings.Contains(content, "<"),
				"has_hooks":      strings.Contains(content, "useState") || strings.Contains(content, "useEffect"),
				"has_components": strings.Contains(content, "Component") || strings.Contains(content, "const "),
			},
		}
		apd.frameworks = append(apd.frameworks, result)
	}

	// Next.js detection
	nextjsConf := apd.detectNextJS(content, packageJSON)
	if nextjsConf > 0.6 {
		result := DetectionResult{
			Type:       "framework",
			Name:       string(NextJSFramework),
			Confidence: nextjsConf,
			Evidence:   apd.getNextJSEvidence(content, packageJSON),
			Location:   filePath,
			Metadata: map[string]interface{}{
				"has_pages_dir":  strings.Contains(filePath, "/pages/"),
				"has_api_routes": strings.Contains(filePath, "/api/"),
				"has_ssr":        strings.Contains(content, "getServerSideProps") || strings.Contains(content, "getStaticProps"),
			},
		}
		apd.frameworks = append(apd.frameworks, result)
	}

	// Vue.js detection
	vueConf := apd.detectVue(content, packageJSON)
	if vueConf > 0.5 {
		result := DetectionResult{
			Type:       "framework",
			Name:       string(VueFramework),
			Confidence: vueConf,
			Evidence:   apd.getVueEvidence(content, packageJSON),
			Location:   filePath,
			Metadata: map[string]interface{}{
				"has_template":    strings.Contains(content, "<template>"),
				"has_script":      strings.Contains(content, "<script>"),
				"has_composition": strings.Contains(content, "setup()") || strings.Contains(content, "ref("),
			},
		}
		apd.frameworks = append(apd.frameworks, result)
	}

	// Express.js detection
	expressConf := apd.detectExpress(content, packageJSON)
	if expressConf > 0.5 {
		result := DetectionResult{
			Type:       "framework",
			Name:       string(ExpressFramework),
			Confidence: expressConf,
			Evidence:   apd.getExpressEvidence(content, packageJSON),
			Location:   filePath,
			Metadata: map[string]interface{}{
				"has_app":        strings.Contains(content, "app.") || strings.Contains(content, "express()"),
				"has_routes":     strings.Contains(content, "app.get") || strings.Contains(content, "app.post"),
				"has_middleware": strings.Contains(content, "app.use"),
			},
		}
		apd.frameworks = append(apd.frameworks, result)
	}

	return nil
}

// detectArchitecturalStyles identifies architectural patterns
func (apd *ArchitecturePatternDetector) detectArchitecturalStyles(filePath, content string) error {
	// MVC Pattern detection
	mvcConf := apd.detectMVC(filePath, content)
	if mvcConf > 0.6 {
		result := DetectionResult{
			Type:       "architectural_style",
			Name:       string(MVCArchitecture),
			Confidence: mvcConf,
			Evidence:   apd.getMVCEvidence(filePath, content),
			Location:   filePath,
			Metadata: map[string]interface{}{
				"has_controllers": strings.Contains(filePath, "controller") || strings.Contains(content, "Controller"),
				"has_models":      strings.Contains(filePath, "model") || strings.Contains(content, "Model"),
				"has_views":       strings.Contains(filePath, "view") || strings.Contains(content, "View"),
			},
		}
		apd.architectural = append(apd.architectural, result)
	}

	// Clean Architecture detection
	cleanConf := apd.detectCleanArchitecture(filePath, content)
	if cleanConf > 0.6 {
		result := DetectionResult{
			Type:       "architectural_style",
			Name:       string(CleanArchitecture),
			Confidence: cleanConf,
			Evidence:   apd.getCleanArchitectureEvidence(filePath, content),
			Location:   filePath,
			Metadata: map[string]interface{}{
				"has_entities":  strings.Contains(filePath, "entities") || strings.Contains(filePath, "domain"),
				"has_use_cases": strings.Contains(filePath, "usecases") || strings.Contains(filePath, "services"),
				"has_adapters":  strings.Contains(filePath, "adapters") || strings.Contains(filePath, "infrastructure"),
			},
		}
		apd.architectural = append(apd.architectural, result)
	}

	// Component-based architecture (common in React)
	componentConf := apd.detectComponentBased(filePath, content)
	if componentConf > 0.7 {
		result := DetectionResult{
			Type:       "architectural_style",
			Name:       string(ComponentBasedArch),
			Confidence: componentConf,
			Evidence:   apd.getComponentBasedEvidence(filePath, content),
			Location:   filePath,
			Metadata: map[string]interface{}{
				"has_components":  strings.Contains(filePath, "components"),
				"has_reusability": strings.Contains(content, "props") || strings.Contains(content, "children"),
				"has_composition": strings.Contains(content, "Component") && strings.Contains(content, "render"),
			},
		}
		apd.architectural = append(apd.architectural, result)
	}

	return nil
}

// detectDesignPatterns identifies common design patterns
func (apd *ArchitecturePatternDetector) detectDesignPatterns(filePath, content string) error {
	// Factory Pattern
	factoryConf := apd.detectFactory(content)
	if factoryConf > 0.7 {
		result := DetectionResult{
			Type:       "design_pattern",
			Name:       string(FactoryPattern),
			Confidence: factoryConf,
			Evidence:   apd.getFactoryEvidence(content),
			Location:   filePath,
			Metadata: map[string]interface{}{
				"has_factory_method": strings.Contains(content, "Factory") || strings.Contains(content, "create"),
				"has_abstraction":    strings.Contains(content, "interface") || strings.Contains(content, "abstract"),
			},
		}
		apd.designPatterns = append(apd.designPatterns, result)
	}

	// Repository Pattern
	repoConf := apd.detectRepository(content)
	if repoConf > 0.7 {
		result := DetectionResult{
			Type:       "design_pattern",
			Name:       string(RepositoryPattern),
			Confidence: repoConf,
			Evidence:   apd.getRepositoryEvidence(content),
			Location:   filePath,
			Metadata: map[string]interface{}{
				"has_repository":  strings.Contains(content, "Repository"),
				"has_crud":        strings.Contains(content, "find") && strings.Contains(content, "create"),
				"has_abstraction": strings.Contains(content, "interface") || strings.Contains(content, "abstract"),
			},
		}
		apd.designPatterns = append(apd.designPatterns, result)
	}

	// Observer Pattern (including React event system)
	observerConf := apd.detectObserver(content)
	if observerConf > 0.6 {
		result := DetectionResult{
			Type:       "design_pattern",
			Name:       string(ObserverPattern),
			Confidence: observerConf,
			Evidence:   apd.getObserverEvidence(content),
			Location:   filePath,
			Metadata: map[string]interface{}{
				"has_events":      strings.Contains(content, "addEventListener") || strings.Contains(content, "onClick"),
				"has_callbacks":   strings.Contains(content, "callback") || strings.Contains(content, "=>"),
				"has_subscribers": strings.Contains(content, "subscribe") || strings.Contains(content, "observer"),
			},
		}
		apd.designPatterns = append(apd.designPatterns, result)
	}

	// Higher Order Component (HOC) Pattern
	hocConf := apd.detectHOC(content)
	if hocConf > 0.7 {
		result := DetectionResult{
			Type:       "design_pattern",
			Name:       string(HOCPattern),
			Confidence: hocConf,
			Evidence:   apd.getHOCEvidence(content),
			Location:   filePath,
			Metadata: map[string]interface{}{
				"has_hoc":         strings.Contains(content, "with") && strings.Contains(content, "Component"),
				"has_wrapper":     strings.Contains(content, "wrapper") || strings.Contains(content, "enhance"),
				"has_composition": strings.Contains(content, "return ") && strings.Contains(content, "Component"),
			},
		}
		apd.designPatterns = append(apd.designPatterns, result)
	}

	// Hooks Pattern
	hooksConf := apd.detectHooksPattern(content)
	if hooksConf > 0.6 {
		result := DetectionResult{
			Type:       "design_pattern",
			Name:       string(HooksPattern),
			Confidence: hooksConf,
			Evidence:   apd.getHooksEvidence(content),
			Location:   filePath,
			Metadata: map[string]interface{}{
				"has_hooks":        strings.Contains(content, "use") && (strings.Contains(content, "State") || strings.Contains(content, "Effect")),
				"has_custom_hooks": strings.Contains(content, "const use") || strings.Contains(content, "function use"),
				"has_lifecycle":    strings.Contains(content, "useEffect") || strings.Contains(content, "useLayoutEffect"),
			},
		}
		apd.designPatterns = append(apd.designPatterns, result)
	}

	return nil
}

// Framework detection methods

func (apd *ArchitecturePatternDetector) detectReact(content, packageJSON string) float64 {
	score := 0.0

	// Package.json evidence
	if strings.Contains(packageJSON, "\"react\"") {
		score += 0.4
	}

	// Import evidence
	if strings.Contains(content, "import React") || strings.Contains(content, "from 'react'") {
		score += 0.3
	}

	// JSX evidence
	if strings.Contains(content, "jsx") || (strings.Contains(content, "<") && strings.Contains(content, ">")) {
		score += 0.2
	}

	// Hooks evidence
	if strings.Contains(content, "useState") || strings.Contains(content, "useEffect") {
		score += 0.1
	}

	return score
}

func (apd *ArchitecturePatternDetector) detectNextJS(content, packageJSON string) float64 {
	score := 0.0

	// Package.json evidence
	if strings.Contains(packageJSON, "\"next\"") {
		score += 0.5
	}

	// Next.js specific imports
	if strings.Contains(content, "next/") {
		score += 0.2
	}

	// SSR/SSG evidence
	if strings.Contains(content, "getServerSideProps") || strings.Contains(content, "getStaticProps") {
		score += 0.2
	}

	// API routes evidence
	if strings.Contains(content, "req.") && strings.Contains(content, "res.") {
		score += 0.1
	}

	return score
}

func (apd *ArchitecturePatternDetector) detectVue(content, packageJSON string) float64 {
	score := 0.0

	// Package.json evidence
	if strings.Contains(packageJSON, "\"vue\"") {
		score += 0.4
	}

	// Vue template evidence
	if strings.Contains(content, "<template>") {
		score += 0.3
	}

	// Vue script evidence
	if strings.Contains(content, "<script>") && strings.Contains(content, "export default") {
		score += 0.2
	}

	// Composition API evidence
	if strings.Contains(content, "ref(") || strings.Contains(content, "reactive(") {
		score += 0.1
	}

	return score
}

func (apd *ArchitecturePatternDetector) detectExpress(content, packageJSON string) float64 {
	score := 0.0

	// Package.json evidence
	if strings.Contains(packageJSON, "\"express\"") {
		score += 0.4
	}

	// Express app evidence
	if strings.Contains(content, "express()") || strings.Contains(content, "app.") {
		score += 0.3
	}

	// Route handler evidence
	if strings.Contains(content, "app.get") || strings.Contains(content, "app.post") {
		score += 0.2
	}

	// Middleware evidence
	if strings.Contains(content, "app.use") {
		score += 0.1
	}

	return score
}

// Architectural style detection methods

func (apd *ArchitecturePatternDetector) detectMVC(filePath, content string) float64 {
	score := 0.0

	// Directory structure evidence
	if strings.Contains(filePath, "controller") || strings.Contains(filePath, "model") || strings.Contains(filePath, "view") {
		score += 0.4
	}

	// Class naming evidence
	if strings.Contains(content, "Controller") || strings.Contains(content, "Model") || strings.Contains(content, "View") {
		score += 0.3
	}

	// Separation pattern evidence
	if strings.Contains(content, "router") && strings.Contains(content, "service") {
		score += 0.2
	}

	// Request/Response handling
	if strings.Contains(content, "req") && strings.Contains(content, "res") {
		score += 0.1
	}

	return score
}

func (apd *ArchitecturePatternDetector) detectCleanArchitecture(filePath, content string) float64 {
	score := 0.0

	// Directory structure evidence
	if strings.Contains(filePath, "domain") || strings.Contains(filePath, "entities") ||
		strings.Contains(filePath, "usecases") || strings.Contains(filePath, "infrastructure") {
		score += 0.4
	}

	// Dependency inversion evidence
	if strings.Contains(content, "interface") || strings.Contains(content, "abstract") {
		score += 0.3
	}

	// Use case evidence
	if strings.Contains(content, "UseCase") || strings.Contains(content, "Service") {
		score += 0.2
	}

	// Repository pattern evidence
	if strings.Contains(content, "Repository") {
		score += 0.1
	}

	return score
}

func (apd *ArchitecturePatternDetector) detectComponentBased(filePath, content string) float64 {
	score := 0.0

	// Component directory evidence
	if strings.Contains(filePath, "components") {
		score += 0.3
	}

	// React component evidence
	if strings.Contains(content, "Component") || strings.Contains(content, "const ") {
		score += 0.3
	}

	// Props evidence (reusability)
	if strings.Contains(content, "props") || strings.Contains(content, "children") {
		score += 0.2
	}

	// Composition evidence
	if strings.Contains(content, "render") || strings.Contains(content, "return") {
		score += 0.2
	}

	return score
}

// Design pattern detection methods

func (apd *ArchitecturePatternDetector) detectFactory(content string) float64 {
	score := 0.0

	// Factory naming evidence
	if strings.Contains(content, "Factory") || strings.Contains(content, "factory") {
		score += 0.4
	}

	// Create method evidence
	if strings.Contains(content, "create") && (strings.Contains(content, "function") || strings.Contains(content, "method")) {
		score += 0.3
	}

	// Abstract creation evidence
	if strings.Contains(content, "interface") || strings.Contains(content, "abstract") {
		score += 0.2
	}

	// Switch/case evidence for object creation
	if strings.Contains(content, "switch") && strings.Contains(content, "case") && strings.Contains(content, "new ") {
		score += 0.1
	}

	return score
}

func (apd *ArchitecturePatternDetector) detectRepository(content string) float64 {
	score := 0.0

	// Repository naming evidence
	if strings.Contains(content, "Repository") {
		score += 0.4
	}

	// CRUD operations evidence
	if strings.Contains(content, "find") && strings.Contains(content, "create") {
		score += 0.3
	}

	// Database abstraction evidence
	if strings.Contains(content, "query") || strings.Contains(content, "database") {
		score += 0.2
	}

	// Interface evidence
	if strings.Contains(content, "interface") {
		score += 0.1
	}

	return score
}

func (apd *ArchitecturePatternDetector) detectObserver(content string) float64 {
	score := 0.0

	// Event listener evidence
	if strings.Contains(content, "addEventListener") || strings.Contains(content, "on") {
		score += 0.3
	}

	// React event evidence
	if strings.Contains(content, "onClick") || strings.Contains(content, "onChange") {
		score += 0.3
	}

	// Observer naming evidence
	if strings.Contains(content, "observer") || strings.Contains(content, "subscribe") {
		score += 0.2
	}

	// Callback evidence
	if strings.Contains(content, "callback") || strings.Contains(content, "=>") {
		score += 0.2
	}

	return score
}

func (apd *ArchitecturePatternDetector) detectHOC(content string) float64 {
	score := 0.0

	// HOC naming convention evidence
	if strings.Contains(content, "with") && strings.Contains(content, "Component") {
		score += 0.4
	}

	// Function returning component evidence
	if strings.Contains(content, "return ") && strings.Contains(content, "Component") {
		score += 0.3
	}

	// Higher-order function evidence
	if strings.Contains(content, "=>") && strings.Contains(content, "=>") { // Nested arrows
		score += 0.2
	}

	// Enhancement evidence
	if strings.Contains(content, "enhance") || strings.Contains(content, "wrapper") {
		score += 0.1
	}

	return score
}

func (apd *ArchitecturePatternDetector) detectHooksPattern(content string) float64 {
	score := 0.0

	// Built-in hooks evidence
	if strings.Contains(content, "useState") || strings.Contains(content, "useEffect") {
		score += 0.4
	}

	// Custom hooks evidence
	if strings.Contains(content, "const use") || strings.Contains(content, "function use") {
		score += 0.3
	}

	// Hook composition evidence
	if strings.Contains(content, "useContext") || strings.Contains(content, "useReducer") {
		score += 0.2
	}

	// Hook rules compliance
	if strings.Contains(content, "useCallback") || strings.Contains(content, "useMemo") {
		score += 0.1
	}

	return score
}

// Evidence extraction methods (return supporting evidence for each detection)

func (apd *ArchitecturePatternDetector) getReactEvidence(content, packageJSON string) []string {
	evidence := make([]string, 0)

	if strings.Contains(packageJSON, "\"react\"") {
		evidence = append(evidence, "React dependency in package.json")
	}
	if strings.Contains(content, "import React") {
		evidence = append(evidence, "React import statement")
	}
	if strings.Contains(content, "jsx") || strings.Contains(content, "<") {
		evidence = append(evidence, "JSX usage detected")
	}
	if strings.Contains(content, "useState") {
		evidence = append(evidence, "React hooks usage")
	}

	return evidence
}

func (apd *ArchitecturePatternDetector) getNextJSEvidence(content, packageJSON string) []string {
	evidence := make([]string, 0)

	if strings.Contains(packageJSON, "\"next\"") {
		evidence = append(evidence, "Next.js dependency in package.json")
	}
	if strings.Contains(content, "next/") {
		evidence = append(evidence, "Next.js imports")
	}
	if strings.Contains(content, "getServerSideProps") {
		evidence = append(evidence, "Server-side rendering")
	}
	if strings.Contains(content, "getStaticProps") {
		evidence = append(evidence, "Static site generation")
	}

	return evidence
}

func (apd *ArchitecturePatternDetector) getVueEvidence(content, packageJSON string) []string {
	evidence := make([]string, 0)

	if strings.Contains(packageJSON, "\"vue\"") {
		evidence = append(evidence, "Vue.js dependency in package.json")
	}
	if strings.Contains(content, "<template>") {
		evidence = append(evidence, "Vue template syntax")
	}
	if strings.Contains(content, "setup()") {
		evidence = append(evidence, "Vue Composition API")
	}

	return evidence
}

func (apd *ArchitecturePatternDetector) getExpressEvidence(content, packageJSON string) []string {
	evidence := make([]string, 0)

	if strings.Contains(packageJSON, "\"express\"") {
		evidence = append(evidence, "Express.js dependency in package.json")
	}
	if strings.Contains(content, "express()") {
		evidence = append(evidence, "Express app initialization")
	}
	if strings.Contains(content, "app.get") {
		evidence = append(evidence, "Express route handlers")
	}
	if strings.Contains(content, "app.use") {
		evidence = append(evidence, "Express middleware")
	}

	return evidence
}

func (apd *ArchitecturePatternDetector) getMVCEvidence(filePath, content string) []string {
	evidence := make([]string, 0)

	if strings.Contains(filePath, "controller") {
		evidence = append(evidence, "Controller directory structure")
	}
	if strings.Contains(content, "Controller") {
		evidence = append(evidence, "Controller class naming")
	}
	if strings.Contains(content, "Model") {
		evidence = append(evidence, "Model class usage")
	}
	if strings.Contains(content, "View") {
		evidence = append(evidence, "View layer implementation")
	}

	return evidence
}

func (apd *ArchitecturePatternDetector) getCleanArchitectureEvidence(filePath, content string) []string {
	evidence := make([]string, 0)

	if strings.Contains(filePath, "domain") {
		evidence = append(evidence, "Domain layer directory")
	}
	if strings.Contains(filePath, "usecases") {
		evidence = append(evidence, "Use cases layer")
	}
	if strings.Contains(content, "interface") {
		evidence = append(evidence, "Interface abstractions")
	}
	if strings.Contains(content, "Repository") {
		evidence = append(evidence, "Repository pattern usage")
	}

	return evidence
}

func (apd *ArchitecturePatternDetector) getComponentBasedEvidence(filePath, content string) []string {
	evidence := make([]string, 0)

	if strings.Contains(filePath, "components") {
		evidence = append(evidence, "Components directory structure")
	}
	if strings.Contains(content, "props") {
		evidence = append(evidence, "Component props usage")
	}
	if strings.Contains(content, "children") {
		evidence = append(evidence, "Component composition")
	}
	if strings.Contains(content, "Component") {
		evidence = append(evidence, "Component class/function")
	}

	return evidence
}

func (apd *ArchitecturePatternDetector) getFactoryEvidence(content string) []string {
	evidence := make([]string, 0)

	if strings.Contains(content, "Factory") {
		evidence = append(evidence, "Factory class naming")
	}
	if strings.Contains(content, "create") {
		evidence = append(evidence, "Factory create method")
	}
	if strings.Contains(content, "switch") && strings.Contains(content, "case") {
		evidence = append(evidence, "Factory switch-case pattern")
	}

	return evidence
}

func (apd *ArchitecturePatternDetector) getRepositoryEvidence(content string) []string {
	evidence := make([]string, 0)

	if strings.Contains(content, "Repository") {
		evidence = append(evidence, "Repository class naming")
	}
	if strings.Contains(content, "find") && strings.Contains(content, "create") {
		evidence = append(evidence, "CRUD operations")
	}
	if strings.Contains(content, "interface") {
		evidence = append(evidence, "Repository interface")
	}

	return evidence
}

func (apd *ArchitecturePatternDetector) getObserverEvidence(content string) []string {
	evidence := make([]string, 0)

	if strings.Contains(content, "addEventListener") {
		evidence = append(evidence, "Event listeners")
	}
	if strings.Contains(content, "onClick") {
		evidence = append(evidence, "React event handlers")
	}
	if strings.Contains(content, "subscribe") {
		evidence = append(evidence, "Observer subscription")
	}

	return evidence
}

func (apd *ArchitecturePatternDetector) getHOCEvidence(content string) []string {
	evidence := make([]string, 0)

	if strings.Contains(content, "with") && strings.Contains(content, "Component") {
		evidence = append(evidence, "HOC naming convention")
	}
	if strings.Contains(content, "return ") && strings.Contains(content, "Component") {
		evidence = append(evidence, "Component returning function")
	}
	if strings.Contains(content, "enhance") {
		evidence = append(evidence, "Component enhancement pattern")
	}

	return evidence
}

func (apd *ArchitecturePatternDetector) getHooksEvidence(content string) []string {
	evidence := make([]string, 0)

	if strings.Contains(content, "useState") {
		evidence = append(evidence, "useState hook usage")
	}
	if strings.Contains(content, "useEffect") {
		evidence = append(evidence, "useEffect hook usage")
	}
	if strings.Contains(content, "const use") {
		evidence = append(evidence, "Custom hook definition")
	}
	if strings.Contains(content, "useContext") {
		evidence = append(evidence, "Context hook usage")
	}

	return evidence
}

// Analysis result methods

func (apd *ArchitecturePatternDetector) GetFrameworks() []DetectionResult {
	return apd.frameworks
}

func (apd *ArchitecturePatternDetector) GetArchitecturalStyles() []DetectionResult {
	return apd.architectural
}

func (apd *ArchitecturePatternDetector) GetDesignPatterns() []DetectionResult {
	return apd.designPatterns
}

func (apd *ArchitecturePatternDetector) GetPrimaryFramework() *DetectionResult {
	if len(apd.frameworks) == 0 {
		return nil
	}

	// Return the framework with highest confidence
	primary := &apd.frameworks[0]
	for i := range apd.frameworks {
		if apd.frameworks[i].Confidence > primary.Confidence {
			primary = &apd.frameworks[i]
		}
	}

	return primary
}

func (apd *ArchitecturePatternDetector) GetPatternComplianceAssessment() map[string]interface{} {
	assessment := make(map[string]interface{})

	// Framework compliance
	primaryFramework := apd.GetPrimaryFramework()
	if primaryFramework != nil {
		assessment["primary_framework"] = primaryFramework.Name
		assessment["framework_confidence"] = primaryFramework.Confidence
		assessment["framework_compliance"] = apd.assessFrameworkCompliance(primaryFramework.Name)
	}

	// Architectural style assessment
	if len(apd.architectural) > 0 {
		assessment["architectural_styles"] = len(apd.architectural)
		assessment["architecture_clarity"] = apd.assessArchitecturalClarity()
	}

	// Design pattern usage
	assessment["design_patterns_count"] = len(apd.designPatterns)
	assessment["pattern_consistency"] = apd.assessPatternConsistency()

	return assessment
}

func (apd *ArchitecturePatternDetector) assessFrameworkCompliance(frameworkName string) string {
	// Simple compliance assessment based on detected patterns
	switch frameworkName {
	case string(ReactFramework):
		if len(apd.designPatterns) > 2 {
			return "good" // Using multiple React patterns
		}
		return "basic"
	case string(NextJSFramework):
		return "good" // Next.js has strong conventions
	default:
		return "unknown"
	}
}

func (apd *ArchitecturePatternDetector) assessArchitecturalClarity() string {
	if len(apd.architectural) == 1 {
		return "clear" // Single architectural style
	} else if len(apd.architectural) <= 2 {
		return "mixed" // Multiple but manageable
	} else {
		return "unclear" // Too many different styles
	}
}

func (apd *ArchitecturePatternDetector) assessPatternConsistency() string {
	if len(apd.designPatterns) <= 3 {
		return "consistent" // Reasonable number of patterns
	} else if len(apd.designPatterns) <= 6 {
		return "moderate" // Many patterns but could be organized
	} else {
		return "inconsistent" // Too many different patterns
	}
}

func (apd *ArchitecturePatternDetector) ExportToJSON() ([]byte, error) {
	result := map[string]interface{}{
		"frameworks":            apd.frameworks,
		"architectural_styles":  apd.architectural,
		"design_patterns":       apd.designPatterns,
		"compliance_assessment": apd.GetPatternComplianceAssessment(),
	}

	return json.MarshalIndent(result, "", "  ")
}
