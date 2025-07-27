# Detailed Explanation of the Bank UI Color System Logic

## 1 Color Psychology and Financial Industry Characteristics

As a global financial service institution, the bank’s UI color logic must meet the following core requirements:
- **Trust and Professionalism**: Financial products need to convey a visual sense of **stability and reliability**, avoiding overly lively or entertainment-style colors.
- **International Inclusiveness**: Consider cultural differences in color perception (e.g., red represents good fortune in the East, but caution or warning in the West).
- **Information Clarity**: In data-intensive financial scenarios, ensure **color contrast and readability**.
- **Brand Consistency**: Continue to use the bank’s classic red (Hex #DB0011) as the core brand color to convey energy and authority.

> Key Insight: Financial products usually use **blue hues** (security, reliability) or **red hues** (energy, trust) as the primary color. The bank chooses the latter to strengthen brand recognition.

## 2 The Core Role of the HSB Color Model

The bank’s color system is based on the **HSB (Hue/Saturation/Brightness)** model, which is superior to the traditional RGB/CMYK models:

| **Dimension** | **Parameter Range** | **Application**        | **Bank Example**       |
|---------------|---------------------|------------------------|-----------------------|
| Hue (H)       | 0°–360°             | Defining primary/secondary colors | Red primary color H=355° |
| Saturation (S)| 0%–100%             | Control of color vividness | Brand red S=100%       |
| Brightness (B)| 0%–100%             | Light and dark hierarchy  | Text gray B=20%        |

**Operating Guidelines**:
- Avoid the lower-right picker area (S > 85%, B < 15%) which can produce muddy colors.
- Apply a **5°–10° hue shift** for neutral colors (e.g., blue-gray) to enhance visual comfort.
- Generate the same color series by adjusting S/B values rather than H values.

## 3 Primary Color System Definition Logic

### 3.1 Brand Primary Color (Bank Red)
```markdown
- HEX: #DB0011
- HSB: (355°, 100%, 86%)
- Usage: Primary buttons, key icons, brand identity
- Color Emotion: Energy, trust, call to action
```

### 3.2 Primary Color Derivation Rules
Use the **line segment coordinate method** to generate color steps:
1. Connect the brand color coordinate (S₀, B₀) to pure white point (0, 100).
2. Connect the brand color coordinate (S₀, B₀) to pure black point (100, 0).
3. Divide each line into 5 equal parts to generate an 11-level color scale.

**Bank Red Series Example**:

| Level    | HSB Value         | Usage Scenario         |
|----------|-------------------|-----------------------|
| Primary  | (355, 100, 86)    | Core call-to-action    |
| Light    | (355, 40, 98)     | Hover state           |
| Dark     | (355, 100, 56)    | Pressed state         |

## 4 Secondary Color System Construction Logic

### 4.1 Hue Wheel Selection Strategy
Based on the primary hue 355°, calculate secondary color positions:
```math
Complementary color = (355° + 180°) % 360 = 175° (teal green)
Adjacent colors = [355° ± 15°, 355° ± 30°] → orange-red / magenta ranges
```

**Actual adopted scheme**:
- **Functional secondary colors**: Avoid direct use of complementary color; select **adjacent colors of the complementary color**.
- **Emotional secondary colors**: Use **analogous colors** (±30°) for harmony.

### 4.2 Secondary Color Calibration Technique
**Visual brightness calibration process**:
1. Choose initial secondary color HSB value (e.g., H=175, S=85, B=90).
2. Overlay with pure black (#000) using **Hue blending mode**.
3. Compare with achromatic brightness and adjust B for consistent visual weight.

**Bank secondary color system**:

| Type          | HSB Value       | Semantic    | Application Scenario    |
|---------------|-----------------|-------------|------------------------|
| Positive Tip  | (120, 85, 80)   | Success     | Transaction complete    |
| Warning Tip   | (45, 90, 95)    | Warning     | Risk alert             |
| Error Tip     | (5, 85, 90)     | Danger      | Operation failed       |
| Information   | (210, 70, 90)   | Notification| System messages        |

## 5 Neutral Color System Design

### 5.1 Desaturation Layer Structure
```markdown
1. **Text Colors** (3–5 steps)
   - Important Titles HSB(0, 0, 20)
   - Body Text HSB(0, 0, 40)
   - Secondary Text HSB(0, 0, 60)

2. **Background Colors**
   - Global Background HSB(220, 5, 98)
   - Card Background HSB(220, 3, 100)

3. **Dividers**
   - Emphasized Divider HSB(0, 0, 90)
   - Regular Divider HSB(0, 0, 95)
```

### 5.2 Advanced Gray Techniques
- Avoid pure black or white: add **brand hue shift** (H=220° blue-gray tone)
- Generation formula: `HSB(220, S<10%, B>90%)`
- Contrast control: text/background brightness difference ≥ 50%.

## 6 Functional Color System Specifications

### 6.1 Status Feedback Colors

| Status    | HSB Value        | Applied Components        |
|-----------|------------------|---------------------------|
| Default   | (355, 100, 86)   | Button/input borders      |
| Hover     | (355, 80, 92)    | Hover effects             |
| Pressed   | (355, 100, 56)   | Click animations          |
| Disabled  | (0, 0, 80)       | Disabled state            |

### 6.2 Data Visualization Palette
Using a **24-color wheel division method**:
1. Base at primary color 355°
2. Sample colors at 15° intervals: `H = (355 + 15 * n) % 360`
3. Adjust S/B to maintain visual balance

**Chart color principles**:
- Key data: use primary color series
- Contrasting data: use complementary colors
- Multi-data sets: limit to **within 5 colors** (primary, analogous, mid-tone colors)

## 7 Practical Application Guidelines

### 7.1 Color Proportion Norms

| Color Type  | Proportion | Usage Description          |
|-------------|------------|----------------------------|
| Primary     | 60–70%     | Dominant visual focus      |
| Secondary   | 20–30%     | Functional prompts/secondary elements |
| Neutral     | 10–20%     | Background/text/dividers   |

### 7.2 Cross-platform Adaptation Solutions
- **Web**: HSL conversion formula `HSL(H, S%, L%)`
- **iOS**: UIColor initialization using sRGB color space
- **Android**: Color.argb() for managing alpha channel

### 7.3 Accessibility Standards
- **Text contrast**: AA level ≥ 4.5:1 (body text) / AAA level ≥ 7:1 (small text)
- **Color blindness adaptation**: Use texture patterns to assist color information (e.g., stripes/dot patterns)
- **Dark mode**: Reduce saturation (S - 30%), increase brightness (B + 15%)

> **Design validation tools**:  
> 1. Ant Design Color Contrast Analyzer  
> 2. Figma Plugins: Stark / A11y  
> 3. WCAG 2.1 Compliance Checker

This system combines **scientific color derivation** (hue wheel angle calculations) with **perceptual calibration** (brightness weighting balance). It ensures brand gene continuity while meeting multiple standards for functionality, accessibility, and internationalization in financial products. Practical use requires design tokens for consistent cross-platform management.