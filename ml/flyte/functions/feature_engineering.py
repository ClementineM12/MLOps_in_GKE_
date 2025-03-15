import cv2
import numpy as np
import functions.glob as flyte_glob


def feature_engineer(
    bucket: str,
    metadata_filename: str,
    data_path: str
):

    dataset = flyte_glob.get_metadata_file(
        bucket=bucket, 
        metadata_filename=metadata_filename, 
        data_path=data_path, 
        target="pickle"
    )
    
    print(f"Example of segmented image array:\n {dataset.iloc[0]['segmented_image']}")
    dataset = calculate_perimeter(dataset)
    dataset = calculate_non_zeros(dataset)
    dataset = calculate_circularity(dataset)
    dataset = calculate_assymetry(dataset)
    dataset = split_channels(dataset)

    flyte_glob.put_metadata_file(
        metadata=dataset, 
        bucket=bucket, 
        metadata_filename=metadata_filename, 
        data_path=data_path
    )
    return
    
    

def calculate_perimeter(images_array):
    images_array['perimeter'] = ''
    for i in range(len(images_array['segmented_image'])):
        # Find contours in the binary image
        image = images_array['segmented_image'][i]
        contours, _ = cv2.findContours(image, cv2.RETR_EXTERNAL, cv2.CHAIN_APPROX_SIMPLE)
        
        # Compute the perimeter of the largest contour
        images_array['perimeter'][i] = cv2.arcLength(contours[0], True) if len(contours) > 0 else 0
    
    return images_array


def calculate_non_zeros(images_array):
    images_array['non_zeros'] = ''
    for i in range(len(images_array['segmented_image'])):

        image = images_array['segmented_image'][i]
        images_array['non_zeros'][i] = np.count_nonzero(image)

    return images_array


def calculate_circularity(images_array):
    images_array['circularity'] = ''
    for i in range(len(images_array['segmented_image'])):
        
        image = images_array['segmented_image'][i]
        
        contours, _ = cv2.findContours(image, cv2.RETR_EXTERNAL, cv2.CHAIN_APPROX_SIMPLE)
        if contours:
            max_contour = max(contours, key=cv2.contourArea)
            perimeter = cv2.arcLength(max_contour, True)
            area = cv2.contourArea(max_contour)
            # Avoid division by zero in case the perimeter is 0
            if perimeter > 0:
                circularity = (4 * np.pi * area) / (perimeter * perimeter)
            else:
                circularity = np.nan
            images_array['circularity'][i] = circularity
        else:
            print(f"No contours found for image {images_array['image_id'][i]}, so circularity remains NaN or you can assign a default value.")
            images_array['circularity'][i] = np.nan
    
    return images_array


def calculate_assymetry(images_array):
    # Initialize new columns with a default value (NaN)
    images_array['main_assymetry'] = np.nan
    images_array['secondary_assymetry'] = np.nan

    for i in range(len(images_array['segmented_image'])):
        image = images_array['segmented_image'][i]
        # Find contours in the binary image
        contours, _ = cv2.findContours(image, cv2.RETR_EXTERNAL, cv2.CHAIN_APPROX_SIMPLE)
        
        # Check if any contours were found
        if not contours:
            # No contours found; keep NaN values or assign a default value if desired
            continue

        # Compute the largest contour
        max_contour = max(contours, key=cv2.contourArea)

        # Compute moments; check if m00 is non-zero to avoid division by zero
        moments = cv2.moments(max_contour)
        if moments["m00"] == 0:
            continue  # Skip this image or leave NaN values

        # Compute centroid
        cx = int(moments["m10"] / moments["m00"])
        cy = int(moments["m01"] / moments["m00"])

        # Compute the central moments
        m11 = moments["m11"] - cx * moments["m01"]
        m20 = moments["m20"] - cx * moments["m10"]
        m02 = moments["m02"] - cy * moments["m01"]

        # Avoid division by zero for asymmetry calculation (denom = m20 + m02)
        if (m20 + m02) == 0:
            continue

        # Compute asymmetry relative to the first axis of symmetry
        theta1 = 0.5 * np.arctan2(2 * m11, m20 - m02)
        main_assymetry = np.abs(m20 * np.sin(theta1)**2 - 2 * m11 * np.sin(theta1) * np.cos(theta1) + m02 * np.cos(theta1)**2) / (m20 + m02)
        images_array['main_assymetry'][i] = main_assymetry

        # Compute asymmetry relative to the second axis of symmetry
        theta2 = 0.5 * np.arctan2(2 * m11, -(m20 - m02))
        secondary_assymetry = np.abs(m20 * np.sin(theta2)**2 - 2 * m11 * np.sin(theta2) * np.cos(theta2) + m02 * np.cos(theta2)**2) / (m20 + m02)
        images_array['secondary_assymetry'][i] = secondary_assymetry

    return images_array

def split_channels(images_array):
    images_array['r_channel'] = ''
    images_array['g_channel'] = ''
    images_array['b_channel'] = ''
    for i in range(len(images_array['image'])):
        
        image = images_array['image'][i]
        image_cropped = image[60:400, 50:550]
        # Use only the pixels that are in the segmented image
        red, green, blue = cv2.split(image_cropped)
        images_array['r_channel'][i] = red
        images_array['g_channel'][i] = green
        images_array['b_channel'][i]  = blue
        
    return images_array
