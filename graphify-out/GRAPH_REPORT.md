# Graph Report - .  (2026-05-17)

## Corpus Check
- Corpus is ~23,108 words - fits in a single context window. You may not need a graph.

## Summary
- 429 nodes · 597 edges · 63 communities (40 shown, 23 thin omitted)
- Extraction: 72% EXTRACTED · 27% INFERRED · 0% AMBIGUOUS · INFERRED: 163 edges (avg confidence: 0.82)
- Token cost: 9,000 input · 3,500 output

## Community Hubs (Navigation)
- [[_COMMUNITY_Go Checker Utilities|Go Checker Utilities]]
- [[_COMMUNITY_Python Build Tools|Python Build Tools]]
- [[_COMMUNITY_Go Agent Core|Go Agent Core]]
- [[_COMMUNITY_Game Item Templates|Game Item Templates]]
- [[_COMMUNITY_Project Documentation|Project Documentation]]
- [[_COMMUNITY_Go Membership System|Go Membership System]]
- [[_COMMUNITY_Go PI Environment|Go PI Environment]]
- [[_COMMUNITY_Python Install Scripts|Python Install Scripts]]
- [[_COMMUNITY_Simulation Room UI|Simulation Room UI]]
- [[_COMMUNITY_Python Validation Tools|Python Validation Tools]]
- [[_COMMUNITY_Windows HDR Display|Windows HDR Display]]
- [[_COMMUNITY_Daily Rewards Pipeline|Daily Rewards Pipeline]]
- [[_COMMUNITY_Go Resource IO|Go Resource I/O]]
- [[_COMMUNITY_Python Install Choreography|Python Install Choreography]]
- [[_COMMUNITY_Go HTTP Networking|Go HTTP Networking]]
- [[_COMMUNITY_Game Skill Levels|Game Skill Levels]]
- [[_COMMUNITY_Python Workspace Setup|Python Workspace Setup]]
- [[_COMMUNITY_Pipeline Maintenance Scripts|Pipeline Maintenance Scripts]]
- [[_COMMUNITY_Go HTTP Handlers|Go HTTP Handlers]]
- [[_COMMUNITY_Python Configuration Tools|Python Configuration Tools]]
- [[_COMMUNITY_Community & Support|Community & Support]]
- [[_COMMUNITY_Shop UI Templates|Shop UI Templates]]
- [[_COMMUNITY_Star Anis Story Stages|Star Anis Story Stages]]
- [[_COMMUNITY_Python Custom Actions|Python Custom Actions]]
- [[_COMMUNITY_Python Custom Recognition|Python Custom Recognition]]
- [[_COMMUNITY_Daily Recruit Operations|Daily Recruit Operations]]
- [[_COMMUNITY_Outpost Management|Outpost Management]]
- [[_COMMUNITY_Arena Code Templates|Arena Code Templates]]
- [[_COMMUNITY_B-Side Idol Event|B-Side Idol Event]]
- [[_COMMUNITY_Windows Process Enumeration|Windows Process Enumeration]]
- [[_COMMUNITY_Common Confirm Buttons|Common Confirm Buttons]]
- [[_COMMUNITY_Star Anis Story 1|Star Anis Story 1]]
- [[_COMMUNITY_Community 41|Community 41]]
- [[_COMMUNITY_Community 42|Community 42]]
- [[_COMMUNITY_Community 43|Community 43]]
- [[_COMMUNITY_Community 44|Community 44]]
- [[_COMMUNITY_Community 45|Community 45]]
- [[_COMMUNITY_Community 46|Community 46]]
- [[_COMMUNITY_Community 47|Community 47]]
- [[_COMMUNITY_Community 48|Community 48]]
- [[_COMMUNITY_Community 49|Community 49]]
- [[_COMMUNITY_Community 50|Community 50]]
- [[_COMMUNITY_Community 51|Community 51]]
- [[_COMMUNITY_Community 52|Community 52]]
- [[_COMMUNITY_Community 53|Community 53]]
- [[_COMMUNITY_Community 54|Community 54]]
- [[_COMMUNITY_Community 55|Community 55]]
- [[_COMMUNITY_Community 56|Community 56]]
- [[_COMMUNITY_Community 57|Community 57]]
- [[_COMMUNITY_Community 60|Community 60]]
- [[_COMMUNITY_Community 61|Community 61]]
- [[_COMMUNITY_Community 62|Community 62]]

## God Nodes (most connected - your core abstractions)
1. `Print()` - 38 edges
2. `t()` - 16 edges
3. `warn()` - 15 edges
4. `main()` - 13 edges
5. `info()` - 11 edges
6. `checkMembership()` - 10 edges
7. `install_maafw()` - 10 edges
8. `Project Agent Instructions` - 10 edges
9. `Get()` - 9 edges
10. `install_mxu()` - 9 edges

## Surprising Connections (you probably didn't know these)
- `Pipeline Node Naming Convention` --conceptually_related_to--> `Pipeline JSONs`  [INFERRED]
  docs/pipeline-node-naming.md → AGENTS.md
- `MDA Project Logo` --references--> `GitHub Install Workflow`  [EXTRACTED]
  assets/logo.png → .github/workflows/install.yml
- `Advise Number 20 Template Image` --conceptually_related_to--> `Advise Collection Completed Visible Pipeline Node`  [AMBIGUOUS]
  assets/resource/image/Advise/20.png → assets/resource/pipeline/Advise/AdviseFlow.json
- `Common Confirm Button Template Image` --conceptually_related_to--> `Common Confirm Action Pipeline Node`  [AMBIGUOUS]
  assets/resource/image/Common/Confirm.png → assets/resource/pipeline/Common/Interaction.json
- `install_deps()` --calls--> `Print()`  [INFERRED]
  tools/install.py → agent/go-service/pkg/maafocus/maafocus.go

## Hyperedges (group relationships)
- **Go Service Warning Templates** — aspect_ratio_warn, hdr_warn, membership_warn, process_warn [INFERRED 0.85]
- **Common Close Button Variants Group** — common_close_button1_template, common_close_button2_template, common_close_button3_template, common_close_button4_template, pipeline_common_close_page [EXTRACTED 1.00]
- **Battle Automation Control Templates Group** — battle_auto_burst_template, battle_auto_shoot_template, pipeline_battle_click_auto_burst, pipeline_battle_click_auto_shoot [EXTRACTED 1.00]
- **Common Go Back Button Variants Group** — common_go_back_template, common_go_back1_template, pipeline_common_go_back [EXTRACTED 1.00]
- **Next Anomaly Button Visual Variants** — Interception_NextAnomaly1, Interception_NextAnomaly2, Interception_NextAnomaly3, Interception_NextAnomaly4, Interception_NextAnomaly5 [INFERRED 0.95]
- **UI Notification Indicator Templates** — Common_RedDot, Common_PassRedDot, Common_New [INFERRED 0.80]
- **Daily Rewards Pass Flow Nodes** — DailyRewards_Pass_DailyRewardsPassStart, DailyRewards_Pass_DailyRewardsClickPassRedDot, DailyRewards_Pass_DailyRewardsEnterPass, DailyRewards_Pass_DailyRewardsClickPassRewardRedDot, DailyRewards_Pass_DailyRewardsPassChangeClick [INFERRED 0.85]
- **Star Anis Event UI Templates** — staranis_logo, staranis_story1_stage, staranis_story1_stage_repeatable, staranis_story2_hard, staranis_story2_hard_repeatable, staranis_story2_hard_sp, staranis_story2_normal, staranis_story2_normal_repeatable [EXTRACTED 1.00]
- **Outpost System UI Templates** — outpost_brief_encounter, outpost_command_center, outpost_daily_lives, outpost_event [EXTRACTED 1.00]
- **Arena Shop UI Templates** — arena_shop, arena_code_manual_selection, arena_electric_manual, arena_fire_manual, arena_iron_manual [EXTRACTED 1.00]
- **Common Shop UI Templates** — commonshopselected_template, commonshopnotselected_template, freereset_template, gemcheck_template, coredustcase_template [INFERRED 0.85]
- **Recycling Shop Item Templates** — goodteamworkbox_template, maintenancekitbox_template, curatedmanufacturerarms_template, gem_template, credit_template, creditcase_template, battledatasetcase_template [INFERRED 0.85]
- **Upgrade Materials Recognition Set** — battledatasetcase_template, creditcase_template, credit_template, coredustcase_template [EXTRACTED 1.00]
- **Shop Voucher Template Images** — shop_voucher_darlingforaday, shop_voucher_elysionvipbuffet, shop_voucher_missilisvipcard, shop_voucher_pilgrimsupplyset, shop_voucher_tetravipspa [INFERRED 0.85]
- **Simulation Room Enemy Buff Up III Templates** — simroom_enemyarmorup3, simroom_enemyvitalsup3, simroom_enemyweaponup3 [INFERRED 0.90]
- **B-Side Idol Event Image Assets** — smallevent_bsideidol_logo, smallevent_bsideidol_stagenormal [INFERRED 0.85]
- **B-SideIdol Event Visual Assets** — bsideidolstagenormalrepeatable_image, bsideidol_stagenormal_image, bsideidol_logo_image [INFERRED 0.80]
- **SmallEvent Quick Battle Stage Click Flow** — smallevent_eventstage_quickbattle_flow, smallevent_story1stage_repeatable_click, smallevent_story2stage_normal_repeatable_click [EXTRACTED 1.00]

## Communities (63 total, 23 thin omitted)

### Community 0 - "Go Checker Utilities"
Cohesion: 0.1
Nodes (24): AspectRatioChecker, aspectRatioRequirement(), buildWarningData(), calculateAspectRatio(), displayController(), displayControllerType(), exactResolutionRequirement(), isAspectRatio16x9() (+16 more)

### Community 1 - "Python Build Tools"
Cohesion: 0.15
Nodes (31): main(), Print(), find_png_images(), main(), optimize_image(), Optimize PNG images in the assets directory using oxipng., Find all PNG images under the given root directory., Optimize a single PNG image using oxipng. Returns True if file was changed. (+23 more)

### Community 2 - "Go Agent Core"
Cohesion: 0.08
Nodes (22): levelFilterWriter, initLogger(), getCwd(), main(), registerAll(), SetVersion(), MyCustomAction, MyCustomRecognition (+14 more)

### Community 3 - "Game Item Templates"
Cohesion: 0.08
Nodes (29): Abnormal Annual Ticket Template, Battle Data Set Case Template, Core Dust Case Template, Credit Currency Icon Template, Credit Case Template, Curated Manufacturer Arms Template, Free Reset Button Template, Gem Item Template (+21 more)

### Community 4 - "Project Documentation"
Cohesion: 0.11
Nodes (28): Project Agent Instructions, Aspect Ratio Warning Template, Common Domain, Interface Description (en_us), Interface Description (zh_cn), Domain+Action+Role Pattern, DoroHelper, Entered Sentinel Pattern (+20 more)

### Community 5 - "Go Membership System"
Cohesion: 0.13
Nodes (20): GenerateDeviceCodeV6(), hashString(), MatchDeviceCodeV6(), queryWMI(), queryWMIFiltered(), queryWMIFirstFixed(), readMachineGuid(), DeviceCodeV6 (+12 more)

### Community 6 - "Go PI Environment"
Cohesion: 0.18
Nodes (16): Controller, Env, ClientLanguage(), ClientMaaFWVersion(), ClientName(), ClientVersion(), ControllerName(), ControllerType() (+8 more)

### Community 7 - "Python Install Scripts"
Cohesion: 0.32
Nodes (11): build_go_agent(), check_go_environment(), configure_ocr_model(), copy_directory(), create_directory_link(), create_file_link(), init_local(), main() (+3 more)

### Community 8 - "Simulation Room UI"
Cohesion: 0.24
Nodes (13): Enemy Buff Up Category, Simulation Room Game Feature, Simulation Room Rules Category, Aggressive Tactics Buff Template, Battle Flag UI Template, Claim Immediately Button Template, Enemy Armor Up III Buff Template, Enemy Vitals Up III Buff Template (+5 more)

### Community 9 - "Python Validation Tools"
Cohesion: 0.27
Nodes (11): create_validator(), find_line_number(), get_validator_class(), load_jsonc(), main(), 在文件中查找JSON路径对应的行号      为了避免找到错误的子字段，只返回顶层对象的行号     例如：/NoSmallGlobe/recognition, 创建 validator，使用新的 referencing API 或回退到 RefResolver, 移除 JSONC 注释，保持 JSON 结构完整 (+3 more)

### Community 10 - "Windows HDR Display"
Cohesion: 0.2
Nodes (10): DISPLAYCONFIG_DEVICE_INFO_HEADER, DISPLAYCONFIG_GET_ADVANCED_COLOR_INFO, DISPLAYCONFIG_MODE_INFO, DISPLAYCONFIG_PATH_INFO, DISPLAYCONFIG_PATH_SOURCE_INFO, DISPLAYCONFIG_PATH_TARGET_INFO, DISPLAYCONFIG_RATIONAL, IsHDREnabled() (+2 more)

### Community 11 - "Daily Rewards Pipeline"
Cohesion: 0.25
Nodes (11): Advise Enter Nikke Page Pipeline Node, Advise Episode Viewing Pipeline Node, New Content Indicator Template, Pass Red Dot Notification Template, Generic Red Dot Notification Template, Daily Rewards Friend Points Start Pipeline Node, Daily Rewards Click Pass Red Dot Pipeline Node, Daily Rewards Click Pass Reward Red Dot Pipeline Node (+3 more)

### Community 12 - "Go Resource I/O"
Cohesion: 0.31
Nodes (6): FindResource(), ReadJsonResource(), ReadResource(), GetResourceBase(), getStandardResourceBase(), resourcePathSink

### Community 13 - "Python Install Choreography"
Cohesion: 0.29
Nodes (5): build_go_agent(), install_agent(), install_deps(), 编译 Go Agent 为目标平台的二进制文件, 安装编译好的 Go Agent 二进制到 install 目录

### Community 14 - "Go HTTP Networking"
Cohesion: 0.25
Nodes (8): Common Close Button Variant 1 Template Image, Common Close Button Variant 2 Template Image, Common Close Button Variant 3 Template Image, Common Close Button Variant 4 Template Image, Common Go Back Button Variant 1 Template Image, Common Go Back Button Template Image, Common Close Page Pipeline Node, Common Go Back Pipeline Node

### Community 15 - "Game Skill Levels"
Cohesion: 0.39
Nodes (8): Daily Rewards Free Recruit Next Recruit Click Pipeline Node, Next Recruit Arrow Template, Interception Next Anomaly Pipeline Node, Next Anomaly Button Template 1, Next Anomaly Button Template 2, Next Anomaly Button Template 3, Next Anomaly Button Template 4, Next Anomaly Button Template 5

### Community 16 - "Python Workspace Setup"
Cohesion: 0.38
Nodes (5): cleanNode(), isPlainObject(), jsonFiles, pipelineRoot, processFile()

### Community 17 - "Pipeline Maintenance Scripts"
Cohesion: 0.48
Nodes (7): Shop Game Feature, Voucher Items, Darling for a Day Voucher Template, Elysion VIP Buffet Voucher Template, Missilis VIP Card Template, Pilgrim Supply Set Voucher Template, Tetra VIP Spa Voucher Template

### Community 18 - "Go HTTP Handlers"
Cohesion: 0.38
Nodes (7): B-SideIdol Logo Template, B-SideIdol Stage Normal Template, B-SideIdol Stage Normal Repeatable Template, SmallEvent Event Stage Quick Battle Flow, SmallEvent Story1 Stage Repeatable Click, SmallEvent Story2 Stage Normal Repeatable Click, SmallEvent Task Configuration

### Community 19 - "Python Configuration Tools"
Cohesion: 0.33
Nodes (6): Maa Pipeline Support VSCode Plugin, MAA Project Reference, oxipng Optimizer, Personalization Configuration, Pre-commit Hooks, Prettier Formatter

### Community 20 - "Community & Support"
Cohesion: 0.53
Nodes (6): Contact Info (en_us), Contact Info (zh_cn), Discord Server, MDA GitHub Issues, MDA GitHub Repository, QQ User Group

### Community 21 - "Shop UI Templates"
Cohesion: 0.4
Nodes (6): Common Shop Not Selected Tab Template, Common Shop Selected Tab Template, Shop Enter Common Shop, Shop Enter Recycling Shop, Recycling Shop Tab Template, Time Flag Template

### Community 22 - "Star Anis Story Stages"
Cohesion: 0.4
Nodes (5): Star Anis Story 2 Hard Stage Template, Star Anis Story 2 Hard Repeatable Template, Star Anis Story 2 Hard SP Stage Template, Star Anis Story 2 Normal Stage Template, Star Anis Story 2 Normal Repeatable Template

### Community 25 - "Daily Recruit Operations"
Cohesion: 0.5
Nodes (4): Coordinated Operations Mode Icon Template, Coordinated Operations Search Pipeline Node, Daily Rewards Free Recruit Visible Pipeline Node, Free Recruit Indicator Template

### Community 26 - "Outpost Management"
Cohesion: 0.5
Nodes (4): Outpost Brief Encounter Template, Outpost Command Center Template, Outpost Daily Lives Template, Outpost Event Template

### Community 27 - "Arena Code Templates"
Cohesion: 1.0
Nodes (4): Arena Code Manual Selection Box Template, Arena Electric Code Manual Template, Arena Fire Code Manual Template, Arena Iron Code Manual Template

### Community 28 - "B-Side Idol Event"
Cohesion: 1.0
Nodes (4): B-Side Idol Event, Small Event Game Feature, B-Side Idol Event Logo Template, B-Side Idol Stage Normal Template

### Community 30 - "Common Confirm Buttons"
Cohesion: 0.67
Nodes (3): Common Confirm Button Template Image, Common Confirm With Circle Button Template Image, Common Confirm Action Pipeline Node

### Community 31 - "Star Anis Story 1"
Cohesion: 0.67
Nodes (3): Star Anis Event Logo Template, Star Anis Story 1 Stage Template, Star Anis Story 1 Stage Repeatable Template

## Ambiguous Edges - Review These
- `Advise Number 20 Template Image` → `Advise Collection Completed Visible Pipeline Node`  [AMBIGUOUS]
  assets/resource/pipeline/Advise/AdviseFlow.json · relation: conceptually_related_to
- `Common Confirm Button Template Image` → `Common Confirm Action Pipeline Node`  [AMBIGUOUS]
  assets/resource/pipeline/Common/Interaction.json · relation: conceptually_related_to

## Knowledge Gaps
- **129 isolated node(s):** `Win32Config`, `Controller`, `Resource`, `Env`, `LUID` (+124 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **23 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **What is the exact relationship between `Advise Number 20 Template Image` and `Advise Collection Completed Visible Pipeline Node`?**
  _Edge tagged AMBIGUOUS (relation: conceptually_related_to) - confidence is low._
- **What is the exact relationship between `Common Confirm Button Template Image` and `Common Confirm Action Pipeline Node`?**
  _Edge tagged AMBIGUOUS (relation: conceptually_related_to) - confidence is low._
- **Why does `Print()` connect `Python Build Tools` to `Go Checker Utilities`, `Go Agent Core`, `Go Membership System`, `Python Install Scripts`, `Python Validation Tools`, `Python Install Choreography`, `Python Custom Actions`?**
  _High betweenness centrality (0.116) - this node is a cross-community bridge._
- **Why does `warn()` connect `Go Checker Utilities` to `Python Build Tools`, `Go Agent Core`, `Go Resource I/O`, `Go Membership System`?**
  _High betweenness centrality (0.105) - this node is a cross-community bridge._
- **Why does `doInit()` connect `Go Checker Utilities` to `Go Agent Core`, `Go PI Environment`?**
  _High betweenness centrality (0.029) - this node is a cross-community bridge._
- **Are the 37 inferred relationships involving `Print()` (e.g. with `main()` and `.run()`) actually correct?**
  _`Print()` has 37 INFERRED edges - model-reasoned connections that need verification._
- **Are the 13 inferred relationships involving `warn()` (e.g. with `main()` and `loadMessages()`) actually correct?**
  _`warn()` has 13 INFERRED edges - model-reasoned connections that need verification._