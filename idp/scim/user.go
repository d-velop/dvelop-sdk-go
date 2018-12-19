// Package scim provides types for representing users and groups using the
// Simple Cloud Identity Management (SCIM) core schema 1.0
//
// cf. http://www.simplecloud.info/specs/draft-scim-core-schema-00.html#schema
package scim

import (
	"encoding/json"
)

// Principal represents a user.
//
// It complies to the SCIM User Schema.
// cf. http://www.simplecloud.info/specs/draft-scim-core-schema-00.html#user-resource
type Principal struct {
	// ID is a unique identifier for the SCIM Resource as defined by the Service Provider.
	//
	// Each representation of the Resource MUST include a non-empty id value. This identifier MUST be unique across the Service Provider's entire set of Resources. It MUST be a stable, non-reassignable identifier that does not change when the same Resource is returned in subsequent requests. The value of the id attribute is always issued by the Service Provider and MUST never be specified by the Service Consumer. bulkId: is a reserved keyword and MUST NOT be used in the unique identifier. REQUIRED and READ-ONLY.
	Id string `json:"id"`

	// ExternalID is a unique identifier for the Resource as defined by the Service Consumer.
	//
	// The externalId may simplify identification of the Resource between Service Consumer and Service provider by allowing the Consumer to refer to the Resource with its own identifier, obviating the need to store a local mapping between the local identifier of the Resource and the identifier used by the Service Provider. Each Resource MAY include a non-empty externalId value. The value of the externalId attribute is always issued be the Service Consumer and can never be specified by the Service Provider. This identifier MUST be unique across the Service Consumer's entire set of Resources. It MUST be a stable, non-reassignable identifier that does not change when the same Resource is returned in subsequent requests. The Service Provider MUST always interpret the externalId as scoped to the Service Consumer's tenant.
	ExternalId string `json:"externalId"`

	// UserName is a unique identifier for the User, typically used by the user to directly authenticate to the service provider.
	//
	// Often displayed to the user as their unique identifier within the system (as opposed to id or externalId, which are generally opaque and not user-friendly identifiers). Each User MUST include a non-empty userName value. This identifier MUST be unique across the Service Consumer's entire set of Users. It MUST be a stable ID that does not change when the same User is returned in subsequent requests. REQUIRED.
	UserName string `json:"userName"`

	// Name contains the components of the User's real name.
	//
	// Providers MAY return just the full name as a single string in the formatted sub-attribute, or they MAY return just the individual component attributes using the other sub-attributes, or they MAY return both. If both variants are returned, they SHOULD be describing the same name, with the formatted name indicating how the component attributes should be combined.
	Name UserName `json:"name"`

	// DisplayName is the name of the User, suitable for display to end-users.
	//
	// Each User returned MAY include a non-empty displayName value. The name SHOULD be the full name of the User being described if known (e.g. Babs Jensen or Ms. Barbara J Jensen, III), but MAY be a username or handle, if that is all that is available (e.g. bjensen). The value provided SHOULD be the primary textual label by which this User is normally displayed by the Service Provider when presenting it to end-users.
	DisplayName string `json:"displayName"`

	// ProfileUrl is a fully qualified URL to a page representing the User's online profile.
	ProfileUrl string `json:"profileUrl"`

	// Title is the user’s title, such as “Vice President.”
	Title string `json:"title"`

	// Emails contains E-mail addresses for the User.
	//
	// The value SHOULD be canonicalized by the Service Provider, e.g. bjensen@example.com instead of bjensen@EXAMPLE.COM. Canonical Type values of work, home, and other.
	Emails []UserValue `json:"emails"`

	// Photos contains URLs of photos of the User.
	//
	// The value SHOULD be a canonicalized URL, and MUST point to an image file (e.g. a GIF, JPEG, or PNG image file) rather than to a web page containing an image. Service Providers MAY return the same image at different sizes, though it is recognized that no standard for describing images of various sizes currently exists. Note that this attribute SHOULD NOT be used to send down arbitrary photos taken by this User, but specifically profile photos of the User suitable for display when describing the User. Instead of the standard Canonical Values for type, this attribute defines the following Canonical Values to represent popular photo sizes: photo, thumbnail.
	Photos []UserValue `json:"photos"`

	// PhoneNumbers are the phone numbers for the User.
	//
	// No canonical value is assumed here. Canonical Type values of work, home, mobile, fax, pager and other.
	PhoneNumbers []UserValue `json:"phoneNumbers"`

	// Groups contains a list of groups that the user belongs to, either thorough direct membership, nested groups, or dynamically calculated.
	//
	// The values are meant to enable expression of common group or role based access control models, although no explicit authorization model is defined. It is intended that the semantics of group membership and any behavior or authorization granted as a result of membership are defined by the Service Provider. The Canonical types "direct" and "indirect" are defined to describe how the group membership was derived. Â Direct group membership indicates the User is directly associated with the group and SHOULD indicate that Consumers may modify membership through the Group Resource. Â Indirect membership indicates User membership is transitive or dynamic and implies that Consumers cannot modify indirect group membership through the Group resource but MAY modify direct group membership through the Group resource which MAY influence indirect memberships. Â If the SCIM Service Provider exposes a Group resource, the value MUST be the "id" attribute of the corresponding Group resources to which the user belongs. Since this attribute is read-only, group membership changes MUST be applied via the Group Resource. READ-ONLY.
	Groups []UserGroup `json:"groups"`
}

func (p Principal) String() string {
	b, _ := json.Marshal(p)
	return string(b)
}

type UserName struct {
	// Formatted is the full name, including all middle names, titles, and suffixes as appropriate, formatted for display (e.g. Ms. Barbara Jane Jensen, III.).
	Formatted string `json:"formatted"`
	// FamilyName ist the family name of the User, or "Last Name" in most Western languages (e.g. Jensen given the full name Ms. Barbara Jane Jensen, III.).
	FamilyName string `json:"familyName"`
	// GivenName is the given name of the User, or "First Name" in most Western languages (e.g. Barbara given the full name Ms. Barbara Jane Jensen, III.).
	GivenName string `json:"givenName"`
	// MiddleName is the middle name(s) of the User (e.g. Jane given the full name Ms. Barbara Jane Jensen, III.).
	MiddleName string `json:"middleName"`
	// HonorificPrefix is the honorific prefix(es) of the User, or "Title" in most Western languages (e.g. Ms. given the full name Ms. Barbara Jane Jensen, III.).
	HonorificPrefix string `json:"honorificPrefix"`
	// HonorificSuffix is the honorific suffix(es) of the User, or "Suffix" in most Western languages (e.g. III. given the full name Ms. Barbara Jane Jensen, III.).
	HonorificSuffix string `json:"honorificSuffix"`
}

type UserValue struct {
	Value string `json:"value"`
}

type UserGroup struct {
	Value   string `json:"value"`
	Display string `json:"display"`
}
