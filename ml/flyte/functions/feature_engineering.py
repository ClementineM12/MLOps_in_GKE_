import cv2
import numpy as np


def feature_engineer(dataset):
    dataset = calculate_perimeter(dataset)
    dataset = calculate_non_zeros(dataset)
    dataset = calculate_circularity(dataset)
    dataset = calculate_assymetry(dataset)
    dataset = split_channels(dataset)
    
    return dataset

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
        
        contours, hierarchy = cv2.findContours(image, cv2.RETR_EXTERNAL, cv2.CHAIN_APPROX_SIMPLE)
        max_contour = max(contours, key=cv2.contourArea)
        perimeter = cv2.arcLength(max_contour, True)
        area = cv2.contourArea(max_contour)
        images_array['circularity'][i] = (4*np.pi*area) / (perimeter*perimeter)
    
    return images_array


def calculate_assymetry(images_array):
    images_array['main_assymetry'] = ''
    images_array['secondary_assymetry'] = ''
    for i in range(len(images_array['segmented_image'])):
        
        image = images_array['segmented_image'][i]
        # Find contours in the binary image
        contours, _ = cv2.findContours(image, cv2.RETR_EXTERNAL, cv2.CHAIN_APPROX_SIMPLE)
        max_contour = max(contours, key=cv2.contourArea)
        
        # Compute the moments of the contour
        moments = cv2.moments(max_contour)

        # Compute the centroid of the contour
        cx = int(moments["m10"] / moments["m00"])
        cy = int(moments["m01"] / moments["m00"])

        # Compute the first and second moments of the contour
        m11 = moments["m11"] - cx * moments["m01"]
        m20 = moments["m20"] - cx * moments["m10"]
        m02 = moments["m02"] - cy * moments["m01"]

        # Compute the degree of asymmetry to the first axis of symmetry
        theta1 = 0.5 * np.arctan2(2 * m11, m20 - m02)
        images_array['main_assymetry'][i] = np.abs(m20 * np.sin(theta1)**2 - 2 * m11 * np.sin(theta1) * np.cos(theta1) + m02 * np.cos(theta1)**2) / (m20 + m02)

        # Compute the degree of asymmetry to the second axis of symmetry
        theta2 = 0.5*np.arctan2(2*m11, -(m20 - m02))
        images_array['secondary_assymetry'][i] = np.abs(m20 * np.sin(theta2)**2 - 2 * m11 * np.sin(theta2) * np.cos(theta2) + m02 * np.cos(theta2)**2) / (m20 + m02)

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