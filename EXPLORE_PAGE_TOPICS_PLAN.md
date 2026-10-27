# Explore Page with Topics - Implementation Plan

## Overview
Build a new "Explore" page feature where users can browse travel packages organized by curated **Topics** (not just categories). The topics are:
- Adventure Trips
- Luxury Escapes
- Cultural Journeys
- Honeymoon Deals
- Family Holidays
- Friends

## Current State Analysis

### Existing Models
- **Package** (`model/catalog/catalog_package.go`) - Has `Category` field (string code)
- **PackageCategory** (`model/base/base_package_category.go`) - Lookup table for package categories (honeymoon, adventure, etc.)

### Existing Services
- **HomeServices** (`service/bxapi/home.go`) - `ListExploreTrips()` with optional category filter
- **PackageServices** (`service/catalog/package.go`) - CRUD for packages by agents

### Current API Endpoints (bxapi)
- `GET /bxapi/home/explore` - List explore trips with optional category filter
- `GET /bxapi/home/deals` - List deals
- `GET /bxapi/home/destinations` - Discover destinations

---

## Proposed Architecture

### 1. Data Models

#### New Model: `ExploreTopic` (model/engage/explore_topic.go)
```go
type ExploreTopic struct {
    ID          uint   `json:"id" gorm:"primarykey"`
    Code        string `json:"code" gorm:"size:64;uniqueIndex;comment:Unique code (adventure, luxury, cultural, honeymoon, family, friends)"`
    Name        string `json:"name" gorm:"size:128;index;comment:Display name"`
    NameAr      string `json:"nameAr" gorm:"size:128;comment:Arabic name"`
    Description string `json:"description" gorm:"size:500;comment:Short description"`
    ImageURL    string `json:"imageUrl" gorm:"size:500;comment:Topic banner/hero image"`
    IconURL     string `json:"iconUrl" gorm:"size:500;comment:Topic icon"`
    SortOrder   int    `json:"sortOrder" gorm:"default:0;comment:Display order"`
    IsActive    bool   `json:"isActive" gorm:"default:true;comment:Visible to users"`
    PackageCount int   `json:"packageCount" gorm:"-;comment:Computed count of packages"`
    
    // Timestamps
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

#### New Model: `PackageTopic` (model/engage/package_topic.go) - Many-to-Many
```go
type PackageTopic struct {
    PackageID uint `json:"packageId" gorm:"primaryKey;index"`
    TopicID   uint `json:"topicId" gorm:"primaryKey;index"`
    AddedBy   uint `json:"addedBy" gorm:"comment:Admin/Agent ID who added"`
    AddedAt   time.Time `json:"addedAt"`
    
    // Relations
    Package *Package     `json:"-" gorm:"foreignKey:PackageID"`
    Topic   *ExploreTopic `json:"-" gorm:"foreignKey:TopicID"`
}
```

### 2. Service Layer

#### New Service: `ExploreTopicServices` (service/engage/explore_topic.go)

**Public (User-facing):**
- `ListTopics()` - Get all active topics with package counts
- `GetTopicPackages(topicCode string, req PageReq)` - Paginated packages for a topic
- `GetTopicDetail(topicCode string)` - Topic metadata + featured packages

**Agent-facing:**
- `AgentGetMyTopics(agentID uint)` - Topics with agent's package counts
- `AgentAddPackageToTopic(agentID, packageID, topicID)` - Associate package with topic
- `AgentRemovePackageFromTopic(agentID, packageID, topicID)` - Remove association

**Admin-facing:**
- `AdminListTopics(req PageReq)` - All topics (including inactive)
- `AdminCreateTopic(req TopicReq)` - Create new topic
- `AdminUpdateTopic(id uint, req TopicReq)` - Update topic
- `AdminDeleteTopic(id uint)` - Soft delete topic
- `AdminReorderTopics(topicOrders []TopicOrderReq)` - Bulk update sort order
- `AdminManageTopicPackages(topicID uint, packageIDs []uint)` - Bulk add/remove packages

### 3. API Endpoints

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/bxapi/explore/topics` | Public | List all active topics with package counts |
| GET | `/bxapi/explore/topics/:code` | Public | Get topic detail + featured packages |
| GET | `/bxapi/explore/topics/:code/packages` | Public | Paginated packages for topic |
| GET | `/bxapi/agent/explore/topics` | Agent | Agent's topics with their package counts |
| POST | `/bxapi/agent/explore/packages/:id/topics` | Agent | Add package to topic(s) |
| DELETE | `/bxapi/agent/explore/packages/:id/topics/:topicId` | Agent | Remove package from topic |
| GET | `/admin/explore/topics` | Admin | List all topics (paginated) |
| POST | `/admin/explore/topics` | Admin | Create topic |
| PUT | `/admin/explore/topics/:id` | Admin | Update topic |
| DELETE | `/admin/explore/topics/:id` | Admin | Delete topic |
| PUT | `/admin/explore/topics/reorder` | Admin | Bulk reorder topics |
| POST | `/admin/explore/topics/:id/packages` | Admin | Bulk manage packages in topic |

### 4. Request/Response Models

#### Request Models (model/bxapi/request/explore.go)
```go
type ExploreTopicReq struct {
    Code        string `json:"code" binding:"required,alphanum,max=64"`
    Name        string `json:"name" binding:"required,max=128"`
    NameAr      string `json:"nameAr" binding:"max=128"`
    Description string `json:"description" binding:"max=500"`
    ImageURL    string `json:"imageUrl" binding:"url,max=500"`
    IconURL     string `json:"iconUrl" binding:"url,max=500"`
    SortOrder   int    `json:"sortOrder"`
    IsActive    bool   `json:"isActive"`
}

type ExploreTopicPageReq struct {
    PageInfo
    IsActive *bool `json:"isActive" form:"isActive"`
}

type TopicOrderReq struct {
    ID       uint `json:"id" binding:"required"`
    SortOrder int  `json:"sortOrder" binding:"required"`
}
```

#### Response Models (model/bxapi/response/explore.go)
```go
type ExploreTopicCard struct {
    ID           uint   `json:"id"`
    Code         string `json:"code"`
    Name         string `json:"name"`
    NameAr       string `json:"nameAr"`
    Description  string `json:"description"`
    ImageURL     string `json:"imageUrl"`
    IconURL      string `json:"iconUrl"`
    PackageCount int    `json:"packageCount"`
}

type ExploreTopicDetail struct {
    ExploreTopicCard
    FeaturedPackages []CustomerExploreTripCard `json:"featuredPackages"`
}

type ExploreTopicPackagesResult struct {
    Topic   ExploreTopicCard          `json:"topic"`
    Packages []CustomerExploreTripCard `json:"packages"`
    Total   int64                     `json:"total"`
    Page    int                       `json:"page"`
    PageSize int                      `json:"pageSize"`
}
```

### 5. Integration Points

#### Package Creation/Update (Agent)
- When agent creates/updates package, optionally associate with topics
- Add `TopicIDs []uint` to `request.PackageReq`

#### Package Listing (Public)
- `ListExploreTrips` can now filter by topic code instead of/in addition to category
- Topics are curated collections; categories are taxonomic classifications

#### Search
- Extend `CustomerSearch` to include topic-based results

---

## Implementation Phases

### Phase 1: Data Models & Migrations (Week 1)
1. Create `ExploreTopic` model in `model/engage/explore_topic.go`
2. Create `PackageTopic` join model in `model/engage/package_topic.go`
3. Create GORM migration (or use auto-migrate)
4. Seed initial 6 topics from requirements

### Phase 2: Service Layer (Week 1-2)
1. Implement `ExploreTopicServices` in `service/engage/explore_topic.go`
2. Add topic association logic to `PackageServices` (CreatePackage, UpdatePackage)
3. Add topic filtering to `HomeServices.ListExploreTrips`

### Phase 3: API Endpoints (Week 2)
1. Add request/response models in `model/bxapi/request/explore.go` and `model/bxapi/response/explore.go`
2. Add routes in router (bxapi for public/agent, admin for admin)
3. Implement handlers in `service/bxapi/explore.go` and `service/system/explore_topic.go`

### Phase 4: Admin Management (Week 2-3)
1. Admin CRUD for topics
2. Bulk package management per topic
3. Drag-and-drop reordering support
4. Topic analytics (package counts, views)

### Phase 5: Frontend Integration (Week 3)
1. Explore page shows topic cards grid
2. Click topic → topic detail page with packages
3. Agent dashboard: package edit → topic multi-select
4. Admin panel: topic management UI

---

## Database Schema

```sql
-- Explore Topics
CREATE TABLE engage_explore_topic (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    code VARCHAR(64) NOT NULL UNIQUE,
    name VARCHAR(128) NOT NULL,
    name_ar VARCHAR(128),
    description VARCHAR(500),
    image_url VARCHAR(500),
    icon_url VARCHAR(500),
    sort_order INT DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_active_sort (is_active, sort_order)
);

-- Package-Topic Association
CREATE TABLE engage_package_topic (
    package_id BIGINT UNSIGNED NOT NULL,
    topic_id BIGINT UNSIGNED NOT NULL,
    added_by BIGINT UNSIGNED NOT NULL,
    added_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (package_id, topic_id),
    FOREIGN KEY (package_id) REFERENCES catalog_package(id) ON DELETE CASCADE,
    FOREIGN KEY (topic_id) REFERENCES engage_explore_topic(id) ON DELETE CASCADE,
    INDEX idx_topic (topic_id)
);
```

---

## Seed Data (Initial Topics)

| Code | Name | NameAr | Description | SortOrder |
|------|------|--------|-------------|-----------|
| adventure | Adventure Trips | رحلات مغامرة | Thrilling experiences for adrenaline seekers | 1 |
| luxury | Luxury Escapes | هروب فاخر | Premium getaways with 5-star amenities | 2 |
| cultural | Cultural Journeys | رحلات ثقافية | Immersive heritage and culture tours | 3 |
| honeymoon | Honeymoon Deals | عروض شهر العسل | Romantic packages for newlyweds | 4 |
| family | Family Holidays | عطلات عائلية | Kid-friendly vacations for all ages | 5 |
| friends | Friends Getaways | رحلات الأصدقاء | Group trips perfect for friends | 6 |

---

## Key Design Decisions

1. **Topics ≠ Categories**: Topics are curated marketing collections; Categories are taxonomic (package.type). A package can belong to multiple topics but has one category.

2. **Agent Control**: Agents can opt-in their packages to topics (subject to admin approval or auto-approval based on package status).

3. **Admin Curation**: Admins control topic visibility, ordering, and can feature packages manually.

4. **Performance**: Package counts cached on topic; use materialized view or counter cache for scale.

5. **Localization**: Topic names support Arabic (`NameAr`) for RTL markets.

---

## Testing Checklist

- [ ] Topic CRUD (admin)
- [ ] Topic listing with counts (public)
- [ ] Topic detail with featured packages
- [ ] Paginated packages by topic
- [ ] Agent can add/remove package from topics
- [ ] Package creation includes topic selection
- [ ] Topic reordering updates display
- [ ] Inactive topics hidden from public
- [ ] Package count updates on association changes
- [ ] Arabic names display correctly

---

## Future Enhancements

1. **Topic Analytics** - Views, clicks, conversion per topic
2. **Seasonal Topics** - Auto-activate/deactivate by date range
3. **Personalized Topics** - ML-based topic recommendations per user
4. **Topic Landing Pages** - SEO-optimized pages per topic
5. **Topic Bundles** - Cross-sell related topics